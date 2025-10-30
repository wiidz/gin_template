package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"gin_template/internal/common/logger"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Request.Header.Get("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}
		c.Set("request_id", id)
		c.Writer.Header().Set("X-Request-ID", id)
		c.Next()
	}
}

func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		ridVal, _ := c.Get("request_id")
		rid, _ := ridVal.(string)
		portVal, _ := c.Get("port")
		port, _ := portVal.(string)
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		fields := []zap.Field{
			zap.String("rid", rid),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.String("ip", c.ClientIP()),
			zap.Int("size", c.Writer.Size()),
		}
		if port != "" {
			fields = append(fields, zap.String("port", port))
		}
		logger.L.Info("access", fields...)
	}
}

func Recovery() gin.HandlerFunc {
	return gin.Recovery()
}

func CORS() gin.HandlerFunc {
	cfg := cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}
	return cors.New(cfg)
}
