package client

import (
	"net/http"

	sagin "github.com/click33/sa-token-go/integrations/gin"
	"github.com/gin-gonic/gin"

	"gin_template/internal/base/repos"
	userhandler "gin_template/internal/domain/client/user"
	usersvc "gin_template/internal/domain/shared/user/service"

	idmng "github.com/wiidz/goutil/mngs/identityMng"
)

func BuildEngine() *gin.Engine {
	e := gin.New()

	// repos.Setup 应在 server/main 处传入
	// 通用仓储直接传入 service
	uRepo := repos.User.Repo
	mng, _ := idmng.NewMng(&idmng.Config{DefaultDevice: "client"})
	uSvc := usersvc.New(uRepo, mng)
	clientH := userhandler.NewClientHandler(uSvc)

	e.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	// 额外业务接口
	v1 := e.Group("/api/v1")
	v1.GET("/ping", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "pong"}) })
	v1.Group("").Use(sagin.CheckLogin()).GET("/user/me", clientH.Me)
	return e
}
