package response

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"gin_template/internal/common/logger"
)

type SuccessResponse[T any] struct {
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

type ErrorResponse struct {
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func JSON(c *gin.Context, status int, data any) {
	c.JSON(status, data)
}

func OK[T any](c *gin.Context, data T) {
	c.JSON(200, SuccessResponse[T]{Msg: "ok", Data: data})
}

func OKMsg[T any](c *gin.Context, msg string, data T) {
	c.JSON(200, SuccessResponse[T]{Msg: msg, Data: data})
}

func Error(c *gin.Context, status int, msg string) {
	// Structured warn log for non-200 responses
	portVal, _ := c.Get("port")
	port, _ := portVal.(string)
	fields := []zap.Field{
		zap.Int("status", status),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
	}
	if port != "" {
		fields = append(fields, zap.String("port", port))
	}
	logger.L.Warn("http_error", fields...)

	c.JSON(status, ErrorResponse{Msg: msg, Data: nil})
	c.Abort()
}
