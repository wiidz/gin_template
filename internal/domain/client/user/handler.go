package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wiidz/goutil/helpers/paramHelper"
	"github.com/wiidz/goutil/structs/networkStruct"

	"gin_template/internal/common/response"
	"gin_template/internal/domain/shared/user/dto"
	usersvc "gin_template/internal/domain/shared/user/service"
)

type ClientHandler struct{ S *usersvc.Service }

func NewClientHandler(s *usersvc.Service) *ClientHandler { return &ClientHandler{S: s} }

func (h *ClientHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := paramHelper.BuildParams(c.Request, &req, networkStruct.BodyJson); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	pair, err := h.S.Login(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error())
		return
	}
	response.OK(c, pair)
}

func (h *ClientHandler) Logout(c *gin.Context) {
	if err := h.S.Logout(c.Request.Context()); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, gin.H{"ok": true})
}

func (h *ClientHandler) Me(c *gin.Context) {
	response.OK(c, gin.H{"login_id": h.S.CurrentLoginID(c.Request.Context())})
}
