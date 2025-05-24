package db

import (
	"context"

	"github.com/google/uuid"
)

// CustomerRepository defines CRUD over customers.
type CustomerRepository interface {
	Create(ctx context.Context, c *Customer) error
	GetByID(ctx context.Context, id uuid.UUID) (*Customer, error)
	List(ctx context.Context, limit, offset int) ([]*Customer, error)
}

// CategoryRepository encapsulates category persistence.
type CategoryRepository interface {
	Create(ctx context.Context, c *Category) error
	GetByID(ctx context.Context, id uuid.UUID) (*Category, error)
	ListChildren(ctx context.Context, parentID *uuid.UUID) ([]*Category, error)
}

// ProductRepository handles products.
type ProductRepository interface {
	Create(ctx context.Context, p *Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*Product, error)
	ListByCategory(ctx context.Context, categoryID uuid.UUID) ([]*Product, error)

	// compute the average price of all products in the subtree of categoryID.
	AveragePriceByCategory(ctx context.Context, categoryID uuid.UUID) (float64, error)
}

// OrderRepository manages orders and items.
type OrderRepository interface {
	CreateOrder(ctx context.Context, o *Order, items []*OrderItem) error
	GetByID(ctx context.Context, id uuid.UUID) (*Order, []*OrderItem, error)
	ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]*Order, error)
}
