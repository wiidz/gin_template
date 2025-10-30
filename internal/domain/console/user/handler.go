package user

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	usersvc "gin_template/internal/domain/shared/user/service"
)

type ConsoleHandler struct{ S *usersvc.Service }

func NewConsoleHandler(s *usersvc.Service) *ConsoleHandler { return &ConsoleHandler{S: s} }

// TODO: implement pagination and filtering when repo supports it
func (h *ConsoleHandler) List(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": []interface{}{}, "total": 0})
}

func (h *ConsoleHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	_, _ = strconv.ParseUint(idStr, 10, 64)
	// TODO: call service when implemented
	c.JSON(http.StatusOK, gin.H{"id": idStr})
}
