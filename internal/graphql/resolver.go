package graphql

import (
	"github.com/felixojiambo/go-graphql-order-service/internal/db"
	"github.com/felixojiambo/go-graphql-order-service/internal/notification"
)

type Resolver struct {
	CategoryRepo    db.CategoryRepository
	ProductRepo     db.ProductRepository
	OrderRepo       db.OrderRepository
	NotificationSvc notification.NotificationService
}

func NewResolver(
	cat db.CategoryRepository,
	prod db.ProductRepository,
	ord db.OrderRepository,
	notif notification.NotificationService,
) *Resolver {
	return &Resolver{cat, prod, ord, notif}
}
