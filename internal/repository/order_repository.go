package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dmehra2102/order-management-platform/internal/domain"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) CreateOrder(ctx context.Context, order *domain.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("start tx: %w", err)
	}

	defer func() {
		_ = tx.Rollback()
	}()

	// Insert order
	_, err = tx.ExecContext(ctx, `
		INSERT INTO orders (
			id, user_id, restaurant_id, total_amount, status,
			created_at, updated_at, version
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`,
		order.ID,
		order.UserID,
		order.RestaurantID,
		order.TotalAmount,
		order.Status,
		order.CreatedAt,
		order.UpdatedAt,
		order.Version,
	)
	if err != nil {
		return fmt.Errorf("insert order: %w", err)
	}

	// Insert items
	for _, item := range order.Items {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO order_items (
				id, order_id, item_id, name, price, quantity, created_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`,
			item.ID,
			order.ID,
			item.ItemID,
			item.Name,
			item.Price,
			item.Quantity,
			order.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("insert order item %s: %w", item.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func (r *OrderRepository) GetOrder(ctx context.Context, orderID string) (*domain.Order, error) {
	order := &domain.Order{}

	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, restaurant_id, total_amount, status, created_at, updated_at, version
		FROM orders WHERE id = $1
	`, orderID).Scan(&order.ID, &order.UserID, &order.RestaurantID, &order.TotalAmount, &order.Status, &order.CreatedAt, &order.UpdatedAt, &order.Version)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found")
		}
		return nil, err
	}

	// Get items
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, item_id, name, price, quantity
		FROM order_items WHERE order_id = $1
	`, orderID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item domain.OrderItem
		if err := rows.Scan(&item.ID, &item.ItemID, &item.Name, &item.Price, &item.Quantity); err != nil {
			return nil, err
		}
		order.Items = append(order.Items, item)
	}

	return order, rows.Err()
}

func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, orderID string, status domain.OrderStatus) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE orders SET status = $1, updated_at = NOW(), version = version + 1
		WHERE id = $2
	`, status, orderID)
	return err
}

func (r *OrderRepository) ListOrders(ctx context.Context, userID string, limit int) ([]domain.Order, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, restaurant_id, total_amount, status, created_at, updated_at, version
		FROM orders WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2
	`, userID, limit)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var order domain.Order
		if err := rows.Scan(&order.ID, &order.UserID, &order.RestaurantID, &order.TotalAmount, &order.Status, &order.CreatedAt, &order.UpdatedAt, &order.Version); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, rows.Err()
}
