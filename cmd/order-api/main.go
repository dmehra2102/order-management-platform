package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/dmehra2102/order-management-platform/internal/config"
	"github.com/dmehra2102/order-management-platform/internal/domain"
	"github.com/dmehra2102/order-management-platform/internal/kafka"
	"github.com/dmehra2102/order-management-platform/internal/logger"
	"github.com/dmehra2102/order-management-platform/internal/metrics"
	"github.com/dmehra2102/order-management-platform/internal/repository"
	"github.com/dmehra2102/order-management-platform/internal/service"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type CreateOrderRequest struct {
	UserID       string                   `json:"user_id"`
	RestaurantID string                   `json:"restaurant_id"`
	Items        []CreateOrderItemRequest `json:"items"`
}

type CreateOrderItemRequest struct {
	ItemID   string  `json:"item_id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

type OrderResponse struct {
	ID           string             `json:"id"`
	UserID       string             `json:"user_id"`
	RestaurantID string             `json:"restaurant_id"`
	Items        []domain.OrderItem `json:"items"`
	TotalAmount  float64            `json:"total_amount"`
	Status       string             `json:"status"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type Server struct {
	mux     *http.ServeMux
	db      *sql.DB
	service *service.OrderService
	logger  *logger.Logger
	metrics *metrics.Metrics
}

func main() {
	cfg := config.Load()
	l := logger.New(cfg.LogLevel)

	l.Info("Starting Order API Service", map[string]any{
		"environment": cfg.Environment,
		"port":        cfg.HTTPPort,
	})

	// Connect to Database
	db, err := sql.Open("postgres", cfg.DatabaseURL())
	if err != nil {
		l.Error("Failed to connect to database", map[string]any{
			"error": err,
		})
		os.Exit(1)
	}
	defer db.Close()

	// Verify DB Connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		l.Error("Failed to pind database", map[string]any{
			"error": err,
		})
		os.Exit(1)
	}

	l.Info("Database connection successful", nil)

	producer := kafka.NewProducer(cfg.KafkaBrokers, l)
	defer producer.Close()

	orderRepo := repository.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepo, producer, l)

	// Prometheus Metrics
	m := metrics.New()
	if err := m.Register(); err != nil {
		l.Warn("Failed to register metrics", map[string]any{
			"error": err,
		})
	}

	// Creating Server
	server := &Server{
		mux:     http.NewServeMux(),
		db:      db,
		service: orderService,
		logger:  l,
		metrics: m,
	}

	server.registerRoutes()

	// HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.HTTPHost, cfg.HTTPPort),
		Handler:      loggingMiddleware(server.mux, l),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// starting http server in background
	go func() {
		l.Info("HTTP Server listening", map[string]any{
			"addr": httpServer.Addr,
		})
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Error("HTTP Server error", map[string]any{
				"error": err,
			})
		}
	}()

	// Shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	l.Info("Shutting down HTTP Server", nil)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		l.Error("HTTP Server shutdown error", map[string]any{
			"error": err,
		})
	}

	l.Info("HTTP Server shutdown complete", nil)
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("POST /api/v1/orders", s.createOrder)
	s.mux.HandleFunc("GET /api/v1/orders/", s.handleGetOrder)
	s.mux.HandleFunc("GET /api/v1/orders", s.listOrders)
	s.mux.Handle("/metrics", promhttp.Handler())
	s.mux.HandleFunc("GET /health", s.healthCheck)
}

func (s *Server) createOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	var items []domain.OrderItem
	for _, item := range req.Items {
		items = append(items, domain.OrderItem{
			ID:       item.ItemID,
			Name:     item.Name,
			Price:    item.Price,
			Quantity: item.Quantity,
		})
	}

	order, err := s.service.CreateOrder(r.Context(), req.UserID, req.RestaurantID, items)
	if err != nil {
		s.respondError(w, http.StatusBadRequest, "Failed to create order", err.Error())
		return
	}

	s.metrics.OrdersCreated.Inc()

	resp := OrderResponse{
		ID:           order.ID,
		UserID:       order.UserID,
		RestaurantID: order.RestaurantID,
		Items:        order.Items,
		TotalAmount:  order.TotalAmount,
		Status:       string(order.Status),
		CreatedAt:    order.CreatedAt,
		UpdatedAt:    order.UpdatedAt,
	}

	s.respondJSON(w, http.StatusCreated, resp)
}

func (s *Server) handleGetOrder(w http.ResponseWriter, r *http.Request) {
	// Extract order ID from /api/v1/orders/{id}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/orders/")
	if path == "" {
		s.respondError(w, http.StatusBadRequest, "Missing order ID", "")
		return
	}

	order, err := s.service.GetOrder(r.Context(), path)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			s.respondError(w, http.StatusNotFound, "Order not found", err.Error())
		} else {
			s.respondError(w, http.StatusInternalServerError, "Failed to fetch order", err.Error())
		}
		return
	}

	resp := OrderResponse{
		ID:           order.ID,
		UserID:       order.UserID,
		RestaurantID: order.RestaurantID,
		Items:        order.Items,
		TotalAmount:  order.TotalAmount,
		Status:       string(order.Status),
		CreatedAt:    order.CreatedAt,
		UpdatedAt:    order.UpdatedAt,
	}

	s.respondJSON(w, http.StatusOK, resp)
}

func (s *Server) listOrders(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		s.respondError(w, http.StatusBadRequest, "Missing user_id query parameter", "")
		return
	}

	orders, err := s.service.ListOrders(r.Context(), userID, 50)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "Failed to list orders", err.Error())
		return
	}

	var responses []OrderResponse
	for _, order := range orders {
		responses = append(responses, OrderResponse{
			ID:           order.ID,
			UserID:       order.UserID,
			RestaurantID: order.RestaurantID,
			Items:        order.Items,
			TotalAmount:  order.TotalAmount,
			Status:       string(order.Status),
			CreatedAt:    order.CreatedAt,
			UpdatedAt:    order.UpdatedAt,
		})
	}

	s.respondJSON(w, http.StatusOK, responses)
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	s.respondJSON(w, http.StatusOK, map[string]any{
		"status": "healthy",
	})
}

func (s *Server) respondJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) respondError(w http.ResponseWriter, code int, message, detail string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	resp := ErrorResponse{
		Error:   message,
		Message: detail,
	}
	json.NewEncoder(w).Encode(resp)
}

func loggingMiddleware(next http.Handler, l *logger.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		l.Debug("HTTP Request", map[string]any{
			"method":   r.Method,
			"path":     r.URL.Path,
			"duration": duration.Milliseconds(),
		})
	})
}
