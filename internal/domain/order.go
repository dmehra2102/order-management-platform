package domain

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusConfirmed OrderStatus = "CONFIRMED"
	OrderStatusFailed    OrderStatus = "FAILED"
	OrderStatusDelivered OrderStatus = "DELIVERED"
	OrderStatusCancelled OrderStatus = "CANCELLED"
)

type OrderItem struct {
	ID       string  `json:"id"`
	ItemID   string  `json:"item_id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

type Order struct {
	ID           string      `json:"id"`
	UserID       string      `json:"user_id"`
	RestaurantID string      `json:"restaurant_id"`
	Items        []OrderItem `json:"items"`
	TotalAmount  float64     `json:"total_amount"`
	Status       OrderStatus `json:"status"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	Version      int         `json:"version"`
}

func NewOrder(userID, restaurantID string, items []OrderItem) (*Order, error) {
	var total float64
	for _, item := range items {
		total += item.Price * float64(item.Quantity)
	}

	now := time.Now().UTC()
	return &Order{
		ID:           uuid.New().String(),
		UserID:       userID,
		RestaurantID: restaurantID,
		Items:        items,
		TotalAmount:  total,
		Status:       OrderStatusPending,
		CreatedAt:    now,
		UpdatedAt:    now,
		Version:      1,
	}, nil
}

func (o *Order) Confirm() {
	o.Status = OrderStatusConfirmed
	o.UpdatedAt = time.Now().UTC() // UTC -> timezone-independent standard
	o.Version++
}

func (o *Order) Fail() {
	o.Status = OrderStatusFailed
	o.UpdatedAt = time.Now().UTC()
	o.Version++
}

func (o *Order) Cancel() {
	o.Status = OrderStatusCancelled
	o.UpdatedAt = time.Now().UTC()
	o.Version++
}
