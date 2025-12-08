package service

import (
	"context"
	"fmt"

	"github.com/dmehra2102/order-management-platform/internal/domain"
	"github.com/dmehra2102/order-management-platform/internal/kafka"
	"github.com/dmehra2102/order-management-platform/internal/logger"
	"github.com/dmehra2102/order-management-platform/internal/repository"
	"github.com/google/uuid"
)

type OrderService struct {
	repo     *repository.OrderRepository
	producer *kafka.Producer
	logger   *logger.Logger
}

func NewOrderService(repo *repository.OrderRepository, producer *kafka.Producer, l *logger.Logger) *OrderService {
	return &OrderService{
		repo:     repo,
		producer: producer,
		logger:   l,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, userID, restaurantID string, items []domain.OrderItem) (*domain.Order, error) {
	if userID == "" || restaurantID == "" {
		return nil, fmt.Errorf("invalid user_id or restaurant_id")
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("order must have at least one item")
	}

	for i := range items {
		if items[i].ID == "" {
			items[i].ID = uuid.New().String()
		}
	}

	// Create Order and save in DB
	order, err := domain.NewOrder(userID, restaurantID, items)
	if err != nil {
		s.logger.Error("Failed to create order", map[string]any{
			"error": err,
		})
		return nil, err
	}

	if err := s.repo.CreateOrder(ctx, order); err != nil {
		s.logger.Error("Failed to save order to database", map[string]any{
			"error":    err,
			"order_id": order.ID,
		})

		return nil, err
	}

	// Publish event
	event := domain.NewOrderCreatedEvent(order)
	if err := s.producer.PublishOrderCreated(ctx, event); err != nil {
		s.logger.Error("Failed to publish order creation event", map[string]any{
			"error":    err,
			"order_id": order.ID,
		})
	}

	return order, nil
}

func (s *OrderService) GetOrder(ctx context.Context, orderID string) (*domain.Order, error) {
	order, err := s.repo.GetOrder(ctx, orderID)
	if err != nil {
		s.logger.Error("Failed to get order", map[string]any{
			"error":    err,
			"order_id": orderID,
		})
		return nil, err
	}
	return order, nil
}

func (s *OrderService) ListOrders(ctx context.Context, userID string, limit int) ([]domain.Order, error) {
	orders, err := s.repo.ListOrders(ctx, userID, limit)
	if err != nil {
		s.logger.Error("Failed to list orders", map[string]interface{}{
			"error":   err,
			"user_id": userID,
		})
		return nil, err
	}
	return orders, nil
}
