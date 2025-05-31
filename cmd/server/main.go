// cmd/server/main.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/mux"

	"github.com/felixojiambo/go-graphql-order-service/internal/auth"
	"github.com/felixojiambo/go-graphql-order-service/internal/db/postgres"
	"github.com/felixojiambo/go-graphql-order-service/internal/graphql"
	"github.com/felixojiambo/go-graphql-order-service/internal/notification"
)

func main() {
	ctx := context.Background()

	// ──────────────────────────────────────────────────────────────────────
	// 1) Initialize Firebase Auth client
	//    Set FIREBASE_SERVICE_ACCOUNT_PATH to the path of your service account JSON,
	//    or leave it empty and rely on GOOGLE_APPLICATION_CREDENTIALS env var.
	serviceAccountPath := os.Getenv("FIREBASE_SERVICE_ACCOUNT_PATH")
	authClient, err := auth.InitializeFirebaseAuthClient(ctx, serviceAccountPath)
	if err != nil {
		log.Fatalf("cannot initialize Firebase Auth client: %v", err)
	}
	// ──────────────────────────────────────────────────────────────────────

	// ──────────────────────────────────────────────────────────────────────
	// 2) Initialize Postgres connection & repositories
	//    Expect DATABASE_URL in the environment, e.g.:
	//      postgres://user:pass@localhost:5432/dbname?sslmode=disable
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL env var is required")
	}
	pgDB, err := postgres.Connect(dbURL)
	if err != nil {
		log.Fatalf("failed to connect to Postgres: %v", err)
	}

	categoryRepo := postgres.NewCategoryRepository(pgDB)
	productRepo := postgres.NewProductRepository(pgDB)
	orderRepo := postgres.NewOrderRepository(pgDB)
	// If you have a CustomerRepository, initialize it here as well:
	// customerRepo := postgres.NewCustomerRepository(pgDB)
	// ──────────────────────────────────────────────────────────────────────

	// ──────────────────────────────────────────────────────────────────────
	// 3) Initialize NotificationService (noop by default)
	//    You can swap in a real SMS/email implementation later.
	noopSvc := &notification.NoopNotificationService{}

	// 4) Construct GraphQL resolver, injecting repos + notification svc
	resolver := graphql.NewResolver(
		categoryRepo,
		productRepo,
		orderRepo,
		noopSvc,
	)
	// ──────────────────────────────────────────────────────────────────────

	// ──────────────────────────────────────────────────────────────────────
	// 5) Create a new gqlgen server pointing to your generated schema
	srv := handler.NewDefaultServer(
		graphql.NewExecutableSchema(graphql.Config{Resolvers: resolver}),
	)
	// ──────────────────────────────────────────────────────────────────────

	// ──────────────────────────────────────────────────────────────────────
	// 6) Set up Gorilla/Mux router and attach Firebase auth middleware
	r := mux.NewRouter()

	// Protect the /query endpoint with FirebaseAuthMiddleware
	r.Handle("/query", auth.FirebaseAuthMiddleware(authClient)(srv))

	// Expose Playground (no auth) on /playground
	r.Handle("/playground", playground.Handler("GraphQL Playground", "/query"))
	// ──────────────────────────────────────────────────────────────────────

	// ──────────────────────────────────────────────────────────────────────
	// 7) Start the HTTP server
	addr := ":8080"
	log.Printf("🚀 Server ready at %s (Playground: http://localhost%s/playground)", addr, addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("http.ListenAndServe failed: %v", err)
	}
	// ──────────────────────────────────────────────────────────────────────
}
