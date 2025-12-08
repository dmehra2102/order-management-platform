package kafka

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/dmehra2102/order-management-platform/internal/domain"
	"github.com/dmehra2102/order-management-platform/internal/logger"
	"github.com/dmehra2102/order-management-platform/internal/repository"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
	logger *logger.Logger
	repo   *repository.OrderRepository
	db     *sql.DB
}

func NewConsumer(brokers, groupID string, l *logger.Logger, repo *repository.OrderRepository, db *sql.DB) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{brokers},
		GroupID:        groupID,
		Topic:          OrdersTopic,
		MinBytes:       10e3,
		MaxBytes:       10e6,
		CommitInterval: 1,
	})

	return &Consumer{
		reader: reader,
		logger: l,
		repo:   repo,
		db:     db,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	c.logger.Info("Consumer started", map[string]any{
		"topic": OrdersTopic,
	})

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			c.logger.Error("Failed to read message", map[string]any{
				"error": err,
			})
			continue
		}

		c.handleOrderCreated(ctx, msg)
	}
}

func (c *Consumer) handleOrderCreated(ctx context.Context, msg kafka.Message) {
	var event domain.OrderCreatedEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		c.logger.Error("Failed to unmarshal OrderCreatedEvent", map[string]any{
			"error": err,
		})
		return
	}

	c.logger.Info("Processing OrderCreatedEvent", map[string]any{
		"order_id":      event.OrderID,
		"user_id":       event.UserID,
		"total_amount":  event.TotalAmount,
		"restaurant_id": event.RestaurantID,
	})

	var newStatus domain.OrderStatus

	if event.TotalAmount < 100 {
		newStatus = domain.OrderStatusFailed
		c.logger.Warn("Order rejected: amount too low", map[string]any{
			"order_id": event.OrderID,
			"amount":   event.TotalAmount,
		})
	} else if event.TotalAmount > 50000 {
		newStatus = domain.OrderStatusFailed
		c.logger.Warn("Order rejected: amount too high", map[string]any{
			"order_id": event.OrderID,
			"amount":   event.TotalAmount,
		})
	} else {
		newStatus = domain.OrderStatusConfirmed
	}

	if err := c.repo.UpdateOrderStatus(ctx, event.OrderID, newStatus); err != nil {
		c.logger.Error("Failed to update order status", map[string]any{
			"error":    err,
			"order_id": event.OrderID,
		})
		return
	}

	// Publishing Event
	producer := NewProducer("localhost:9092,localhost:9094", c.logger)
	defer producer.Close()

	if newStatus == domain.OrderStatusConfirmed {
		confirmEvent := domain.NewOrderConfirmedEvent(event.OrderID)
		if err := producer.PublishOrderConfirmed(ctx, confirmEvent); err != nil {
			c.logger.Error("Failed to publish confirmation", map[string]any{
				"error":    err,
				"order_id": event.OrderID,
			})
		}
	} else {
		failEvent := domain.NewOrderFailedEvent(event.OrderID, "Validation Failed")
		if err := producer.PublishedOrderFailed(ctx, failEvent); err != nil {
			c.logger.Error("Failed to publish failure", map[string]any{
				"error":    err,
				"order_id": event.OrderID,
			})
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
