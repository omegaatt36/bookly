package engine

import (
	"context"
	"net/http"
)

// ContextKey is a custom type for context keys
type ContextKey string

// Context keys
const (
	// ContextKeyUserID is the key for the UserID in the context
	ContextKeyUserID ContextKey = "userID"
)

func getUserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value(ContextKeyUserID).(string); ok {
		return userID
	}

	return ""
}

// WithUserID adds the UserID to the context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ContextKeyUserID, userID)
}

// Context represents a context.
type Context struct {
	Request *http.Request
	userID  string
	// add other context fields as needed
}

// GetUserID returns the user ID from the context
func (c *Context) GetUserID() string {
	return c.userID
}

// SetUserID sets the user ID in the context
func (c *Context) SetUserID(userID string) {
	c.userID = userID
}
