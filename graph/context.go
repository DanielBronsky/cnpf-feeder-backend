package graph

import (
	"context"

	"github.com/gin-gonic/gin"
)

// GinContextKey is the context key for Gin context (SA1029: avoid string keys)
type ginContextKey struct{}

var GinContextKey = ginContextKey{}

// GetGinContext extracts Gin context from GraphQL context
func GetGinContext(ctx context.Context) *gin.Context {
	if gc, ok := ctx.Value(GinContextKey).(*gin.Context); ok {
		return gc
	}
	return nil
}
