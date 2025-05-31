package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

// ErrNoTokenHeader is returned when no Authorization header is present.
var ErrNoTokenHeader = fmt.Errorf("authorization header missing")

// ErrInvalidToken is returned when the provided token is invalid/expired.
var ErrInvalidToken = fmt.Errorf("invalid or expired Firebase ID token")

// AuthVerifier abstracts VerifyIDToken, so we can mock it in tests.
type AuthVerifier interface {
	VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error)
}

// FirebaseAuthMiddleware returns an HTTP middleware that validates Firebase ID tokens.
// On success, it stores *auth.Claims in the request context. On failure, it responds 401.
func FirebaseAuthMiddleware(verifier AuthVerifier) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1) Extract "Authorization: Bearer <ID_TOKEN>"
			header := r.Header.Get("Authorization")
			if header == "" {
				http.Error(w, ErrNoTokenHeader.Error(), http.StatusUnauthorized)
				return
			}
			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, ErrInvalidToken.Error(), http.StatusUnauthorized)
				return
			}
			rawIDToken := parts[1]

			// 2) Verify the token with Firebase
			ctx := r.Context()
			decodedToken, err := verifier.VerifyIDToken(ctx, rawIDToken)
			if err != nil {
				http.Error(w, ErrInvalidToken.Error(), http.StatusUnauthorized)
				return
			}

			// 3) Map Firebase token claims → our *auth.Claims
			c := &Claims{UID: decodedToken.UID}
			if email, ok := decodedToken.Claims["email"].(string); ok {
				c.Email = email
			}
			if rolesIface, ok := decodedToken.Claims["roles"].([]interface{}); ok {
				for _, ri := range rolesIface {
					if rs, ok := ri.(string); ok {
						c.Roles = append(c.Roles, rs)
					}
				}
			}

			// 4) Store claims in context
			ctxWithClaims := NewContext(ctx, c)

			// 5) Call next handler with updated context
			next.ServeHTTP(w, r.WithContext(ctxWithClaims))
		})
	}
}

// InitializeFirebaseAuthClient initializes a Firebase Auth client using the service account JSON.
// If serviceAccountPath is empty, it falls back to GOOGLE_APPLICATION_CREDENTIALS env var.
func InitializeFirebaseAuthClient(ctx context.Context, serviceAccountPath string) (*auth.Client, error) {
	var app *firebase.App
	var err error

	if serviceAccountPath != "" {
		opt := option.WithCredentialsFile(serviceAccountPath)
		app, err = firebase.NewApp(ctx, nil, opt)
		if err != nil {
			return nil, fmt.Errorf("firebase.NewApp: %w", err)
		}
	} else {
		app, err = firebase.NewApp(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("firebase.NewApp (default): %w", err)
		}
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("app.Auth: %w", err)
	}
	return authClient, nil
}
