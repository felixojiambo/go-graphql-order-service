package postgres

import (
	"context"
	"database/sql"
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

// Create inserts a new product.
func (r *productRepo) Create(ctx context.Context, p *db.Product) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO products (id,name,description,price,category_id)
		 VALUES ($1,$2,$3,$4,$5)`,
		p.ID, p.Name, p.Description, p.Price, p.CategoryID,
	)
	return err
}

// GetByID fetches one product.
func (r *productRepo) GetByID(ctx context.Context, id uuid.UUID) (*db.Product, error) {
	var p db.Product
	if err := r.db.GetContext(ctx, &p,
		`SELECT id,name,description,price,category_id FROM products WHERE id=$1`, id,
	); err != nil {
		return nil, err
	}
	return &p, nil
}

// ListByCategory returns all products in a category subtree.
func (r *productRepo) ListByCategory(ctx context.Context, categoryID uuid.UUID) ([]*db.Product, error) {
	// for simplicity assume subtree logic in SQL
	rows, err := r.db.QueryxContext(ctx, `
WITH RECURSIVE ch(id) AS (
    SELECT id FROM categories WHERE id = $1
  UNION ALL
    SELECT c.id FROM categories c
    JOIN ch ON c.parent_id = ch.id
)
SELECT p.id,p.name,p.description,p.price,p.category_id
  FROM products p
  JOIN ch ON p.category_id = ch.id
`, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*db.Product
	for rows.Next() {
		var p db.Product
		if err := rows.StructScan(&p); err != nil {
			return nil, err
		}
		out = append(out, &p)
	}
	return out, nil
}

// AveragePriceByCategory computes the average price over the same subtree.
func (r *productRepo) AveragePriceByCategory(ctx context.Context, categoryID uuid.UUID) (float64, error) {
	var avg sql.NullFloat64
	query := `
WITH RECURSIVE ch(id) AS (
    SELECT id FROM categories WHERE id = $1
  UNION ALL
    SELECT c.id FROM categories c
    JOIN ch ON c.parent_id = ch.id
)
SELECT AVG(p.price) FROM products p
  JOIN ch ON p.category_id = ch.id;
`
	if err := r.db.GetContext(ctx, &avg, query, categoryID); err != nil {
		return 0, fmt.Errorf("avg query: %w", err)
	}
	if !avg.Valid {
		return 0, nil
	}
	return avg.Float64, nil
}
