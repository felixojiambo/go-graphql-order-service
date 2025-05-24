package graphql

import "github.com/felixojiambo/go-graphql-order-service/internal/db"

// Resolver serves as the dependency injector for all your GraphQL resolvers.
// Add any repositories or services this server requires here.
type Resolver struct {
	CategoryRepo db.CategoryRepository
	ProductRepo  db.ProductRepository
}

// NewResolver constructs a Resolver with the required dependencies.
func NewResolver(
	categoryRepo db.CategoryRepository,
	productRepo db.ProductRepository,
) *Resolver {
	return &Resolver{
		CategoryRepo: categoryRepo,
		ProductRepo:  productRepo,
	}
}
