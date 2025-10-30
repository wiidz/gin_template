package console

import (
	"context"
	"net/http"

	sagin "github.com/click33/sa-token-go/integrations/gin"
	"github.com/gin-gonic/gin"

	"gin_template/internal/base/repos"
	userhandler "gin_template/internal/domain/console/user"
	usersvc "gin_template/internal/domain/shared/user/service"

	idmng "github.com/wiidz/goutil/mngs/identityMng"
)

func BuildEngine() *gin.Engine {
	e := gin.New()

	// user console 业务路由（使用 Manager 实例）
	mng, _ := idmng.NewMng(&idmng.Config{DefaultDevice: "client"})
	// repos.Setup 应在 server/main 处传入
	uRepo := repos.User.Repo
	uSvc := usersvc.New(uRepo, &authAdapter{mng: mng})
	uConsole := userhandler.NewConsoleHandler(uSvc)

	e.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	v1 := e.Group("/api/v1")
	{
		v1.GET("/ping", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "pong"}) })
		protected := v1.Group("")
		protected.Use(sagin.CheckLogin(), sagin.CheckRole("admin"))

		protected.GET("/users", uConsole.List)
		protected.GET("/users/:id", uConsole.Get)

		// 这里不再挂载 IAM Subject（已移除复杂 identityMng）
	}
	return e
}

// 与 client 复用相同适配器
type authAdapter struct{ mng *idmng.IdentityMng }

func (a *authAdapter) Login(ctx context.Context, loginId, device string) (string, string, error) {
	pair, err := a.mng.Login(ctx, loginId, device)
	if err != nil {
		return "", "", err
	}
	return pair.AccessToken, pair.RefreshToken, nil
}
func (a *authAdapter) Logout(ctx context.Context) error          { return a.mng.Logout(ctx) }
func (a *authAdapter) CurrentLoginID(ctx context.Context) string { return a.mng.CurrentLoginID(ctx) }
