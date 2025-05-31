// internal/graphql/schema.resolvers.go
package graphql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/felixojiambo/go-graphql-order-service/internal/auth"
	"github.com/felixojiambo/go-graphql-order-service/internal/db"
	"github.com/google/uuid"
)

// Children resolves immediate subcategories for a Category.
// This is unprotected (any authenticated user can list the category tree).
func (r *categoryResolver) Children(ctx context.Context, obj *Category) ([]*Category, error) {
	// (Optional) You could check a role here if you needed to restrict “viewing” the tree.
	// e.g. if !auth.HasRole(ctx, "viewer") { return nil, errors.New("unauthorized") }

	cid, err := uuid.Parse(obj.ID)
	if err != nil {
		return nil, err
	}
	subs, err := r.CategoryRepo.ListChildren(ctx, &cid)
	if err != nil {
		return nil, err
	}
	out := make([]*Category, len(subs))
	for i, sub := range subs {
		out[i] = &Category{
			ID:     sub.ID.String(),
			Name:   sub.Name,
			Parent: &Category{ID: obj.ID},
		}
	}
	return out, nil
}

// CreateCategory persists a new category.
// Only users with the “admin” role may create a category.
func (r *mutationResolver) CreateCategory(ctx context.Context, input NewCategory) (*Category, error) {
	if !auth.HasRole(ctx, "admin") {
		return nil, errors.New("unauthorized: must have 'admin' role to create categories")
	}

	var parentID *uuid.UUID
	if input.ParentID != nil {
		pid, err := uuid.Parse(*input.ParentID)
		if err != nil {
			return nil, errors.New("invalid parentID")
		}
		parentID = &pid
	}

	dbCat := &db.Category{
		ID:       uuid.New(),
		Name:     input.Name,
		ParentID: parentID,
	}
	if err := r.CategoryRepo.Create(ctx, dbCat); err != nil {
		return nil, err
	}

	gqlCat := &Category{
		ID:   dbCat.ID.String(),
		Name: dbCat.Name,
	}
	if parentID != nil {
		gqlCat.Parent = &Category{ID: parentID.String()}
	}
	return gqlCat, nil
}

// CreateProduct persists a new product.
// Only users with the “admin” role may create products.
func (r *mutationResolver) CreateProduct(ctx context.Context, input NewProduct) (*Product, error) {
	if !auth.HasRole(ctx, "admin") {
		return nil, errors.New("unauthorized: must have 'admin' role to create products")
	}

	catID, err := uuid.Parse(input.CategoryID)
	if err != nil {
		return nil, errors.New("invalid categoryID")
	}

	prod := &db.Product{
		ID:          uuid.New(),
		Name:        input.Name,
		Description: input.Description,
		Price:       input.Price,
		CategoryID:  catID,
	}
	if err := r.ProductRepo.Create(ctx, prod); err != nil {
		return nil, err
	}

	return &Product{
		ID:          prod.ID.String(),
		Name:        prod.Name,
		Description: prod.Description,
		Price:       prod.Price,
		Category:    &Category{ID: prod.CategoryID.String()},
	}, nil
}

// Categories returns all root categories.
// Any authenticated user can call this.
func (r *queryResolver) Categories(ctx context.Context) ([]*Category, error) {
	// (Optional) you could require a particular role here
	dbCats, err := r.CategoryRepo.ListChildren(ctx, nil)
	if err != nil {
		return nil, err
	}
	out := make([]*Category, len(dbCats))
	for i, c := range dbCats {
		out[i] = &Category{
			ID:   c.ID.String(),
			Name: c.Name,
		}
	}
	return out, nil
}

// ProductsByCategory returns all products in a category subtree.
// Any authenticated user can call this.
func (r *queryResolver) ProductsByCategory(ctx context.Context, categoryID string) ([]*Product, error) {
	// (Optional) require a role if needed
	cid, err := uuid.Parse(categoryID)
	if err != nil {
		return nil, errors.New("invalid categoryID")
	}
	prods, err := r.ProductRepo.ListByCategory(ctx, cid)
	if err != nil {
		return nil, err
	}
	out := make([]*Product, len(prods))
	for i, p := range prods {
		out[i] = &Product{
			ID:          p.ID.String(),
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Category:    &Category{ID: p.CategoryID.String()},
		}
	}
	return out, nil
}

// AveragePriceByCategory returns the average price of all products in a category subtree.
// Any authenticated user can call this (if you want, you could restrict to "analyst" role, etc).
func (r *queryResolver) AveragePriceByCategory(ctx context.Context, categoryID string) (float64, error) {
	// (Optional) if !auth.HasRole(ctx, "analyst") { return 0, errors.New("unauthorized") }
	cid, err := uuid.Parse(categoryID)
	if err != nil {
		return 0, errors.New("invalid categoryID")
	}
	return r.ProductRepo.AveragePriceByCategory(ctx, cid)
}

// PlaceOrder is the resolver for the placeOrder field.
// Only users with the “customer” role may place orders.
func (r *mutationResolver) PlaceOrder(ctx context.Context, input OrderInput) (*Order, error) {
	if !auth.HasRole(ctx, "customer") {
		return nil, errors.New("unauthorized: must have 'customer' role to place orders")
	}

	// 1) parse & validate customerID
	custID, err := uuid.Parse(input.CustomerID)
	if err != nil {
		return nil, errors.New("invalid customerID")
	}

	// 2) build domain Order + OrderItems
	order := &db.Order{
		ID:         uuid.New(),
		CustomerID: custID,
		CreatedAt:  time.Now(),
	}
	var items []*db.OrderItem
	var total float64

	for _, in := range input.Items {
		pid, err := uuid.Parse(in.ProductID)
		if err != nil {
			return nil, fmt.Errorf("invalid productID %q", in.ProductID)
		}
		// TODO: lookup real price from DB
		price := 0.0

		oi := &db.OrderItem{
			ID:        uuid.New(),
			OrderID:   order.ID,
			ProductID: pid,
			Quantity:  in.Quantity,
			UnitPrice: price, // ← was Price
		}
		total += price * float64(in.Quantity)
		items = append(items, oi)
	}
	order.Total = total

	// 3) persist in a transaction
	if err := r.OrderRepo.CreateOrder(ctx, order, items); err != nil {
		return nil, err
	}

	// 4) fire‐and‐forget notifications
	smsMsg := fmt.Sprintf("Your order %s has been placed. Total: %.2f", order.ID, order.Total)
	go r.NotificationSvc.SendOrderSMS(ctx, "<customer‐phone>", smsMsg)

	emailBody := fmt.Sprintf(
		"Dear customer,\n\nYour order %s for %.2f was successful!",
		order.ID, order.Total,
	)
	go r.NotificationSvc.SendOrderEmail(ctx, "<customer‐email>", "Order Confirmation", emailBody)

	// 5) map back to GraphQL types
	gqlItems := make([]*OrderItem, len(items))
	for i, it := range items {
		gqlItems[i] = &OrderItem{
			ID:       it.ID.String(),
			Product:  &Product{ID: it.ProductID.String()},
			Quantity: it.Quantity,
			Price:    it.UnitPrice, // ← pull from UnitPrice
		}
	}

	return &Order{
		ID:         order.ID.String(),
		CustomerID: order.CustomerID.String(),
		Items:      gqlItems,
		Total:      order.Total,
		CreatedAt:  order.CreatedAt, // time.Time matches your Time scalar
	}, nil
}

// Schema glue — do not edit:
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }
func (r *Resolver) Query() QueryResolver       { return &queryResolver{r} }
func (r *Resolver) Category() CategoryResolver { return &categoryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type categoryResolver struct{ *Resolver }
