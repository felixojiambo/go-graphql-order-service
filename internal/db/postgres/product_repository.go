// internal/db/postgres/product_repository.go
package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/felixojiambo/go-graphql-order-service/internal/db"
)

type productRepo struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) db.ProductRepository {
	return &productRepo{db: db}
}

func (r *productRepo) Create(ctx context.Context, p *db.Product) error {
	const q = `
	INSERT INTO products(id, name, description, price, category_id)
	VALUES ($1, $2, $3, $4, $5)
	`
	if _, err := r.db.ExecContext(ctx, q,
		p.ID, p.Name, p.Description, p.Price, p.CategoryID,
	); err != nil {
		return fmt.Errorf("insert product: %w", err)
	}
	return nil
}

func (r *productRepo) GetByID(ctx context.Context, id uuid.UUID) (*db.Product, error) {
	const q = `
	SELECT id, name, description, price, category_id
	  FROM products
	 WHERE id = $1
	`
	var p db.Product
	if err := r.db.GetContext(ctx, &p, q, id); err != nil {
		return nil, fmt.Errorf("get product by id: %w", err)
	}
	return &p, nil
}

func (r *productRepo) ListByCategory(ctx context.Context, categoryID uuid.UUID) ([]*db.Product, error) {
	const q = `
	WITH RECURSIVE cat_hierarchy AS (
	  SELECT id FROM categories WHERE id = $1
	  UNION ALL
	  SELECT c.id
	    FROM categories c
	    JOIN cat_hierarchy ch ON c.parent_id = ch.id
	)
	SELECT p.id, p.name, p.description, p.price, p.category_id
	  FROM products p
	  JOIN cat_hierarchy ch ON p.category_id = ch.id
	`
	var out []*db.Product
	if err := r.db.SelectContext(ctx, &out, q, categoryID); err != nil {
		return nil, fmt.Errorf("list products by category: %w", err)
	}
	return out, nil
}

func (r *productRepo) AveragePriceByCategory(ctx context.Context, categoryID uuid.UUID) (float64, error) {
	const q = `
	WITH RECURSIVE cat_hierarchy AS (
	  SELECT id FROM categories WHERE id = $1
	  UNION ALL
	  SELECT c.id
	    FROM categories c
	    JOIN cat_hierarchy ch ON c.parent_id = ch.id
	)
	SELECT AVG(p.price)
	  FROM products p
	  JOIN cat_hierarchy ch ON p.category_id = ch.id
	`
	var avg *float64
	if err := r.db.GetContext(ctx, &avg, q, categoryID); err != nil {
		return 0, fmt.Errorf("avg price query: %w", err)
	}
	if avg == nil {
		return 0, nil
	}
	return *avg, nil
}
