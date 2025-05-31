package auth

import "context"

// ContextKey is the type we use for storing values in context.Context
type ContextKey string

const (
	// ContextUserClaimsKey is used to store/retrieve *Claims in request context
	ContextUserClaimsKey ContextKey = "userClaims"
)

// Claims represents the pieces of information we care about from Firebase’s ID token.
type Claims struct {
	UID   string   // Firebase UID (unique user ID)
	Email string   // The user's email, if present
	Roles []string // Custom "roles" array from token (e.g. ["admin","user"])
	// You can extend this struct with any other custom claim fields you need.
}

// NewContext returns a new context that carries the provided *Claims.
func NewContext(parent context.Context, c *Claims) context.Context {
	return context.WithValue(parent, ContextUserClaimsKey, c)
}

// FromContext retrieves *Claims from ctx. Returns (nil,false) if not present.
func FromContext(ctx context.Context) (*Claims, bool) {
	raw := ctx.Value(ContextUserClaimsKey)
	if raw == nil {
		return nil, false
	}
	c, ok := raw.(*Claims)
	return c, ok
}

// HasRole returns true if ctx’s Claims include the given role.
func HasRole(ctx context.Context, role string) bool {
	c, ok := FromContext(ctx)
	if !ok {
		return false
	}
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}
