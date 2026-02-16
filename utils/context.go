package utils

import (
	"context"

	"github.com/cnpf/feeder-backend/graph"
	"github.com/gin-gonic/gin"
)

// GetGinContext extracts Gin context from GraphQL context
func GetGinContext(ctx context.Context) *gin.Context {
	return graph.GetGinContext(ctx)
}
