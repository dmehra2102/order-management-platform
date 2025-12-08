package domain

import (
	"time"

	"github.com/google/uuid"
)

type EventType string

const (
	OrderCreatedEventType   EventType = "OrderCreated"
	OrderConfirmedEventType EventType = "OrderConfirmed"
	OrderFailedEventType    EventType = "OrderFailed"
	OrderCancelledEventType EventType = "OrderCancelled"
)

type Event interface {
	AggregateID() string
	EventType() EventType
	Timestamp() time.Time
}

type OrderCreatedEvent struct {
	EventID      string      `json:"event_id"`
	OrderID      string      `json:"order_id"`
	UserID       string      `json:"user_id"`
	RestaurantID string      `json:"restaurant_id"`
	Items        []OrderItem `json:"items"`
	TotalAmount  float64     `json:"total_amount"`
	CreatedAt    time.Time   `json:"created_at"`
}

func (e OrderCreatedEvent) AggregateID() string  { return e.OrderID }
func (e OrderCreatedEvent) EventType() EventType { return OrderCreatedEventType }
func (e OrderCreatedEvent) Timestamp() time.Time { return e.CreatedAt }

type OrderConfirmedEvent struct {
	EventID     string    `json:"event_id"`
	OrderID     string    `json:"order_id"`
	ConfirmedAt time.Time `json:"confirmed_at"`
}

func (e OrderConfirmedEvent) AggregateID() string  { return e.OrderID }
func (e OrderConfirmedEvent) EventType() EventType { return OrderConfirmedEventType }
func (e OrderConfirmedEvent) Timestamp() time.Time { return e.ConfirmedAt }

type OrderFailedEvent struct {
	EventID  string    `json:"event_id"`
	OrderID  string    `json:"order_id"`
	Reason   string    `json:"reason"`
	FailedAt time.Time `json:"failed_at"`
}

func (e OrderFailedEvent) AggregateID() string  { return e.OrderID }
func (e OrderFailedEvent) EventType() EventType { return OrderFailedEventType }
func (e OrderFailedEvent) Timestamp() time.Time { return e.FailedAt }

func NewOrderCreatedEvent(order *Order) OrderCreatedEvent {
	return OrderCreatedEvent{
		EventID:      uuid.New().String(),
		OrderID:      order.ID,
		UserID:       order.UserID,
		RestaurantID: order.RestaurantID,
		Items:        order.Items,
		TotalAmount:  order.TotalAmount,
		CreatedAt:    order.CreatedAt,
	}
}

func NewOrderConfirmedEvent(orderID string) OrderConfirmedEvent {
	return OrderConfirmedEvent{
		EventID:     uuid.New().String(),
		OrderID:     orderID,
		ConfirmedAt: time.Now().UTC(),
	}
}

func NewOrderFailedEvent(orderID, reason string) OrderFailedEvent {
	return OrderFailedEvent{
		EventID:  uuid.New().String(),
		OrderID:  orderID,
		Reason:   reason,
		FailedAt: time.Now().UTC(),
	}
}
