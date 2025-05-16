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

func getUserIDFromContext(ctx context.Context) int32 {
	if userID, ok := ctx.Value(ContextKeyUserID).(int32); ok {
		return userID
	}

	return 0
}

// WithUserID adds the UserID to the context
func WithUserID(ctx context.Context, userID int32) context.Context {
	return context.WithValue(ctx, ContextKeyUserID, userID)
}

// Context represents a context.
type Context struct {
	Request *http.Request
	userID  int32
}

// GetUserID returns the user ID from the context
func (c *Context) GetUserID() int32 {
	return c.userID
}

// SetUserID sets the user ID in the context
func (c *Context) SetUserID(userID int32) {
	c.userID = userID
}
