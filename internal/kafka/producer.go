package kafka

import (
	"context"
	"encoding/json"

	"github.com/dmehra2102/order-management-platform/internal/domain"
	"github.com/dmehra2102/order-management-platform/internal/logger"
	"github.com/segmentio/kafka-go"
)

const (
	OrdersTopic      = "orders"
	OrderStatusTopic = "order-status"
)

type Producer struct {
	writer *kafka.Writer
	logger *logger.Logger
}

func NewProducer(brokers string, l *logger.Logger) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers),
		RequiredAcks: kafka.RequireAll,
		Balancer:     &kafka.Hash{},
		Compression:  kafka.Snappy,
	}

	return &Producer{
		writer: writer,
		logger: l,
	}
}

func (p *Producer) PublishOrderCreated(ctx context.Context, event domain.OrderCreatedEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		p.logger.Error("Failed to marshal OrderCreatedEvent", map[string]any{"error": err})

		return err
	}

	msg := kafka.Message{
		Topic: OrdersTopic,
		Key:   []byte(event.OrderID),
		Value: payload,
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		p.logger.Error("Failed to publish OrderCreatedEvent", map[string]any{
			"error":    err,
			"order_id": event.OrderID,
		})
		return err
	}

	p.logger.Info("OrderCreatedEvent published", map[string]any{
		"order_id": event.OrderID,
		"user_id":  event.UserID,
	})

	return nil
}

func (p *Producer) PublishOrderConfirmed(ctx context.Context, event domain.OrderConfirmedEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Topic: OrderStatusTopic,
		Key:   []byte(event.OrderID),
		Value: payload,
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		p.logger.Error("Failed to publish OrderConfirmedEvent", map[string]any{
			"error":    err,
			"order_id": event.OrderID,
		})
		return err
	}

	p.logger.Info("OrderConfirmedEvent published", map[string]any{
		"order_id": event.OrderID,
	})

	return nil
}

func (p *Producer) PublishedOrderFailed(ctx context.Context, event domain.OrderFailedEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Topic: OrderStatusTopic,
		Key:   []byte(event.OrderID),
		Value: payload,
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		p.logger.Error("Failed to publish OrderFailedEvent", map[string]any{
			"error":    err,
			"order_id": event.OrderID,
		})

		return err
	}

	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
