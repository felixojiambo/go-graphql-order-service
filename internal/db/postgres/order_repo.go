// internal/postgres/order_repository.go
package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/felixojiambo/go-graphql-order-service/internal/db"
)

type orderRepo struct {
	db *sqlx.DB
}

// NewOrderRepository returns a db.OrderRepository backed by Postgres.
func NewOrderRepository(db *sqlx.DB) db.OrderRepository {
	return &orderRepo{db: db}
}

// CreateOrder inserts an Order and its OrderItems in a single transaction.
func (r *orderRepo) CreateOrder(ctx context.Context, o *db.Order, items []*db.OrderItem) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// insert order
	const insertOrder = `
		INSERT INTO orders (id, customer_id, total_amount, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
	`
	if _, err := tx.ExecContext(
		ctx, insertOrder,
		o.ID, o.CustomerID, o.Total, o.Status,
	); err != nil {
		return fmt.Errorf("insert order: %w", err)
	}

	// insert items
	const insertItem = `
		INSERT INTO order_items (
			id, order_id, product_id, quantity, unit_price, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
	`
	for _, it := range items {
		if _, err := tx.ExecContext(
			ctx, insertItem,
			it.ID, o.ID, it.ProductID, it.Quantity, it.UnitPrice,
		); err != nil {
			return fmt.Errorf("insert order_item %s: %w", it.ID, err)
		}
	}

	return tx.Commit()
}

// GetByID fetches one Order together with its items.
func (r *orderRepo) GetByID(ctx context.Context, id uuid.UUID) (*db.Order, []*db.OrderItem, error) {
	var o db.Order
	const selOrder = `
		SELECT id, customer_id, total_amount, status, created_at, updated_at
		FROM orders
		WHERE id = $1
	`
	if err := r.db.GetContext(ctx, &o, selOrder, id); err != nil {
		return nil, nil, fmt.Errorf("select order: %w", err)
	}

	var items []*db.OrderItem
	const selItems = `
		SELECT id, order_id, product_id, quantity, unit_price, created_at, updated_at
		FROM order_items
		WHERE order_id = $1
		ORDER BY created_at
	`
	if err := r.db.SelectContext(ctx, &items, selItems, id); err != nil {
		return &o, nil, fmt.Errorf("select order_items: %w", err)
	}

	return &o, items, nil
}

// ListByCustomer returns all Orders for a given customer.
func (r *orderRepo) ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]*db.Order, error) {
	var orders []*db.Order
	const sel = `
		SELECT id, customer_id, total_amount, status, created_at, updated_at
		FROM orders
		WHERE customer_id = $1
		ORDER BY created_at DESC
	`
	if err := r.db.SelectContext(ctx, &orders, sel, customerID); err != nil {
		return nil, fmt.Errorf("select orders by customer: %w", err)
	}
	return orders, nil
}
