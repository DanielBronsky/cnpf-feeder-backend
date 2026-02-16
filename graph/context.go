package graph

import (
	"context"

	"github.com/gin-gonic/gin"
)

// GetGinContext extracts Gin context from GraphQL context
func GetGinContext(ctx context.Context) *gin.Context {
	if gc, ok := ctx.Value("ginContext").(*gin.Context); ok {
		return gc
	}
	return nil
}
