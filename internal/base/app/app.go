package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/wiidz/gin_template/internal/base/config"

	appMng "github.com/wiidz/goutil/mngs/appMng"
	"github.com/wiidz/goutil/structs/configStruct"
)

type Instance struct {
	App     *appMng.AppMng
	Project *HTTPProjectConfig
}

type HTTPProjectConfig struct {
	name string
	ip   string
	port string
	addr string
}

func newHTTPProjectConfig(name string) *HTTPProjectConfig {
	return &HTTPProjectConfig{name: name}
}

func (p *HTTPProjectConfig) Build(base *configStruct.BaseConfig) error {
	ip, port := resolveEndpoint(p.name)
	p.ip = ip
	p.port = port
	p.addr = fmt.Sprintf("%s:%s", ip, port)
	return nil
}

func (p *HTTPProjectConfig) Name() string { return p.name }
func (p *HTTPProjectConfig) IP() string   { return p.ip }
func (p *HTTPProjectConfig) Port() string { return p.port }
func (p *HTTPProjectConfig) Addr() string { return p.addr }

var (
	once        sync.Once
	loadErr     error
	clientInst  *Instance
	consoleInst *Instance
)

func Init(ctx context.Context) error {
	once.Do(func() {
		manager := appMng.DefaultManager()
		debug := !isProd(config.C.Env)

		clientLoader := baseLoader("client", debug)
		clientInst = &Instance{Project: newHTTPProjectConfig("client")}
		clientInst.App, loadErr = manager.Get(ctx, appMng.Options{
			ID:            "client",
			Loader:        clientLoader,
			ProjectConfig: clientInst.Project,
			CheckStart: &configStruct.CheckStart{
				Postgres: config.C.DB.DSN != "",
			},
		})
		if loadErr != nil {
			return
		}

		consoleLoader := baseLoader("console", debug)
		consoleInst = &Instance{Project: newHTTPProjectConfig("console")}
		consoleInst.App, loadErr = manager.Get(ctx, appMng.Options{
			ID:            "console",
			Loader:        consoleLoader,
			ProjectConfig: consoleInst.Project,
			CheckStart:    &configStruct.CheckStart{},
		})
		if loadErr != nil {
			return
		}

		consoleInst.App.Postgres = clientInst.App.Postgres
	})
	return loadErr
}

func Client(ctx context.Context) (*Instance, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := Init(ctx); err != nil {
		return nil, err
	}
	return clientInst, nil
}

func Console(ctx context.Context) (*Instance, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := Init(ctx); err != nil {
		return nil, err
	}
	return consoleInst, nil
}

func baseLoader(name string, debug bool) appMng.Loader {
	return appMng.LoaderFunc(func(ctx context.Context) (*appMng.Result, error) {
		ip, port := resolveEndpoint(name)
		base := &configStruct.BaseConfig{
			Profile: &configStruct.AppProfile{
				Name:  name,
				Host:  ip,
				Port:  port,
				Debug: debug,
			},
			Location: time.Local,
		}
		if config.C.DB.DSN != "" {
			base.PostgresConfig = &configStruct.PostgresConfig{DSN: config.C.DB.DSN}
		}
		return &appMng.Result{BaseConfig: base}, nil
	})
}

func resolveEndpoint(name string) (ip, port string) {
	switch name {
	case "client":
		ip = config.C.HTTP2.Client.IP
		port = config.C.HTTP2.Client.Port
	case "console":
		ip = config.C.HTTP2.Console.IP
		port = config.C.HTTP2.Console.Port
	default:
		ip = config.C.HTTP.IP
		port = config.C.HTTP.Port
	}
	if ip == "" {
		ip = config.C.HTTP.IP
	}
	if port == "" {
		if name == "console" {
			port = "8081"
		} else {
			port = config.C.HTTP.Port
		}
	}
	if ip == "" {
		ip = "0.0.0.0"
	}
	if port == "" {
		port = "8080"
	}
	return
}

func isProd(env string) bool {
	switch env {
	case "prod", "production":
		return true
	default:
		return false
	}
}
