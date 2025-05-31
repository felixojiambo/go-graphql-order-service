package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // Postgres driver

	"github.com/felixojiambo/go-graphql-order-service/internal/db"
)

// Connect opens a sqlx.DB to a Postgres instance using the given DSN.
// Example DSN: "postgres://user:pass@localhost:5432/dbname?sslmode=disable"
func Connect(dsn string) (*sqlx.DB, error) {
	dbx, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres.Connect: %w", err)
	}
	return dbx, nil
}

// ----------------------------------------------------------------------
// categoryRepo implements db.CategoryRepository.
// ----------------------------------------------------------------------

type categoryRepo struct {
	db *sqlx.DB
}

func NewCategoryRepository(db *sqlx.DB) db.CategoryRepository {
	return &categoryRepo{db: db}
}

// Create inserts a new category row.
func (r *categoryRepo) Create(ctx context.Context, c *db.Category) error {
	const query = `
		INSERT INTO categories (id, name, parent_id)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.ExecContext(ctx, query, c.ID, c.Name, c.ParentID)
	return err
}

// GetByID fetches one category by its UUID.
func (r *categoryRepo) GetByID(ctx context.Context, id uuid.UUID) (*db.Category, error) {
	var c db.Category
	const query = `
		SELECT id, name, parent_id
		FROM categories
		WHERE id = $1
	`
	if err := r.db.GetContext(ctx, &c, query, id); err != nil {
		return nil, err
	}
	return &c, nil
}

// ListChildren returns all categories whose parent_id = parentID.
// If parentID is nil, returns only the “root” categories (parent_id IS NULL).
func (r *categoryRepo) ListChildren(ctx context.Context, parentID *uuid.UUID) ([]*db.Category, error) {
	var rows []*db.Category
	var err error

	if parentID == nil {
		// Root categories (parent_id IS NULL)
		const query = `
			SELECT id, name, parent_id
			FROM categories
			WHERE parent_id IS NULL
		`
		err = r.db.SelectContext(ctx, &rows, query)
	} else {
		// Immediate children of a given parent
		const query = `
			SELECT id, name, parent_id
			FROM categories
			WHERE parent_id = $1
		`
		err = r.db.SelectContext(ctx, &rows, query, *parentID)
	}

	if err != nil {
		return nil, err
	}
	return rows, nil
}
