package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"gin_template/internal/base/app"
	"gin_template/internal/base/config"
	"gin_template/internal/base/repos"
	"gin_template/internal/base/server"
	"gin_template/internal/common/logger"
)

func main() {
	config.Init()
	logger.Init(config.C.Env)
	defer logger.Sync()

	if config.C.Env == "prod" || config.C.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx := context.Background()
	clientInstance, err := app.Client(ctx)
	if err != nil {
		log.Fatalf("client app init failed: %v", err)
	}
	consoleInstance, err := app.Console(ctx)
	if err != nil {
		log.Fatalf("console app init failed: %v", err)
	}

	if clientInstance.App.Postgres != nil {
		repos.Setup(clientInstance.App.Postgres)
	} else {
		log.Printf("warning: postgres manager not initialized; repositories skipped")
	}

	clientAddr := overrideAddr(clientInstance.Project.IP(), clientInstance.Project.Port(), os.Getenv("PORT_CLIENT"))
	consoleAddr := overrideAddr(consoleInstance.Project.IP(), consoleInstance.Project.Port(), os.Getenv("PORT_CONSOLE"))

	srv := server.NewServer()

	go func() {
		if err := srv.Start(clientAddr, consoleAddr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server start error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
}

func overrideAddr(defaultIP, defaultPort, overridePort string) string {
	ip := defaultIP
	if ip == "" {
		ip = "0.0.0.0"
	}
	port := defaultPort
	if overridePort != "" {
		port = overridePort
	}
	if port == "" {
		port = "8080"
	}
	return fmt.Sprintf("%s:%s", ip, port)
}
