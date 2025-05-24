package db

import (
	"time"

	"github.com/google/uuid"
)

// Customer represents a purchaser in the system.
type Customer struct {
	ID        uuid.UUID `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// Category models a hierarchical product grouping.
type Category struct {
	ID        uuid.UUID  `db:"id"`
	Name      string     `db:"name"`
	ParentID  *uuid.UUID `db:"parent_id"` // nil for root
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
}

// Product defines an item for sale.
type Product struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	Description *string   `db:"description"`
	Price       float64   `db:"price"`
	CategoryID  uuid.UUID `db:"category_id"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// Order represents a customer purchase.
type Order struct {
	ID         uuid.UUID `db:"id"`
	CustomerID uuid.UUID `db:"customer_id"`
	Total      float64   `db:"total_amount"`
	Status     string    `db:"status"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

// OrderItem links products to an order.
type OrderItem struct {
	ID        uuid.UUID `db:"id"`
	OrderID   uuid.UUID `db:"order_id"`
	ProductID uuid.UUID `db:"product_id"`
	Quantity  int       `db:"quantity"`
	UnitPrice float64   `db:"unit_price"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
