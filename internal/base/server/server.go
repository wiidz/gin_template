package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"gin_template/internal/common/middleware"
	clientport "gin_template/internal/domain/client"
	consoleport "gin_template/internal/domain/console"
)

type Server struct {
	clientEngine  *gin.Engine
	consoleEngine *gin.Engine
	clientServer  *http.Server
	consoleServer *http.Server
}

func NewServer() *Server {
	// 1) 初始化 gin
	log.Printf("boot: init gin")

	// 2) 构建路由（client）
	log.Printf("boot: build client engine")
	clientEngine := buildWithRoutePrefix("client", clientport.BuildEngine)
	clientEngine.Use(func(c *gin.Context) { c.Set("port", "client"); c.Next() })
	clientEngine.Use(middleware.IPDenylist())
	clientEngine.Use(middleware.RateLimit(100, 200))
	clientEngine.Use(middleware.RateLimitIP(50, 100))
	sdecorate(clientEngine)

	// 3) 构建路由（console）
	log.Printf("boot: build console engine")
	consoleEngine := buildWithRoutePrefix("console", consoleport.BuildEngine)
	consoleEngine.Use(func(c *gin.Context) { c.Set("port", "console"); c.Next() })
	consoleEngine.Use(middleware.IPDenylist())
	consoleEngine.Use(middleware.RateLimit(100, 200))
	consoleEngine.Use(middleware.RateLimitIP(50, 100))
	sdecorate(consoleEngine)

	return &Server{clientEngine: clientEngine, consoleEngine: consoleEngine}
}

func buildWithRoutePrefix(prefix string, builder func() *gin.Engine) *gin.Engine {
	prev := gin.DebugPrintRouteFunc
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		// Aligned columns: method(6), path(48)
		fmt.Fprintf(gin.DefaultWriter, "[GIN-debug] [%s] %-6s %-48s --> %s (%d handlers)\n",
			strings.ToUpper(prefix), httpMethod, absolutePath, handlerName, nuHandlers)
	}
	fmt.Fprintf(gin.DefaultWriter, "[GIN-debug] ===== %s ROUTES =====\n", strings.ToUpper(prefix))
	e := builder()
	gin.DebugPrintRouteFunc = prev
	return e
}

func sdecorate(e *gin.Engine) {
	// Structured logs (zap)
	e.Use(
		middleware.RequestID(),
		middleware.AccessLog(),
		middleware.Recovery(),
		middleware.CORS(),
	)
}

func (s *Server) Start(clientAddr, consoleAddr string) error {
	s.clientServer = &http.Server{Addr: clientAddr, Handler: s.clientEngine, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second}
	s.consoleServer = &http.Server{Addr: consoleAddr, Handler: s.consoleEngine, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second}

	errCh := make(chan error, 2)
	go func() {
		log.Printf("client listening on %s", clientAddr)
		errCh <- s.clientServer.ListenAndServe()
	}()
	go func() {
		log.Printf("console listening on %s", consoleAddr)
		errCh <- s.consoleServer.ListenAndServe()
	}()

	// return on first error (or nil only if both closed by Shutdown)
	if err := <-errCh; err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.clientServer != nil {
		_ = s.clientServer.Shutdown(ctx)
	}
	if s.consoleServer != nil {
		_ = s.consoleServer.Shutdown(ctx)
	}
	return nil
}
