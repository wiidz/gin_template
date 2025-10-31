## Gin Template

Opinionated, elegant Gin template with config, logging, middlewares, multi-port routing, identity (authn + subject), and graceful shutdown.

go env -w 'GOPRIVATE=github.com/wiidz/*'

### Quick start

1. Copy repo
2. Adjust `module` in `go.mod`
3. Run:

```bash
make tidy
make run
```

### Bootstrap new project

Use the helper script to copy this template and rewrite the module path quickly:

```bash
scripts/new_project.sh my_service --module github.com/you/my_service
```

- 默认目标目录会在模板的同级目录下创建，也可以通过 `--dir /path/to/target` 自定义。
- 如果只传 `my_service`，模块名也会替换成 `my_service`；推荐显式传入完整的 Go Module 路径（`--module`）。
- 脚本会自动执行 `go mod tidy` 并初始化 Git 仓库，可通过 `--skip-tidy` / `--skip-git` 关闭。

### Hot reload (Air)

```bash
go install github.com/cosmtrek/air@latest
make dev
```

Air watches `cmd/`, `internal/`, and config files via `.air.toml` and rebuilds automatically.

### Structure

```
cmd/server                 # Entrypoint; starts client + console HTTP servers
internal/base/app          # App manager wiring (client / console instances)
internal/base/config       # Viper config
internal/base/repos        # Repository registry and DB wiring
internal/base/server       # HTTP servers bootstrap (client + console)
internal/common/logger     # Zap logger
internal/common/middleware # RequestID, AccessLog, Recovery, CORS
internal/common/response   # Response helpers
internal/platform/db       # GORM init, AutoMigrate, WithTx
identity 功能已迁移到 goutil/mngs/identityMng（ginext 一键挂载）。
  authn/                   # Authentication (login/logout/session via Sa-Token stputil)
  subject/                 # Subject directory (USER/STAFF/...) with entity/model/repo/service/handler
  facade/                  # Unified login facade (ensure Subject + issue Token)
internal/domain/client/    # Client HTTP adapter (routes + handlers)
internal/domain/console/   # Console HTTP adapter (routes + handlers)
internal/domain/shared/user/ # Shared user domain (entity/model/repo/service)

### 应用管理（App Manager）

- 使用 `github.com/wiidz/goutil/mngs/appMng` 作为统一的应用实例容器
- `internal/base/app` 封装了 client / console 两个实例的初始化逻辑
- 默认按 `configs/config.yaml` 中的 `http2` 配置推导监听地址，也可以通过 `PORT_CLIENT`、`PORT_CONSOLE` 环境变量覆盖端口
- 数据库连接由 App Manager 统一装载，当前示例采用 PostgreSQL DSN（`config.C.DB.DSN`）
- `cmd/server/main.go` 会在启动时调用 `app.Client` / `app.Console` 并注册仓储（`internal/base/repos`）
```

### Config

- File: `configs/config.yaml`
- Env overrides with `ENV_VAR` matching keys (dots replaced by underscore)
- `.env` is optional via `godotenv`

Supports dual ports:

```
http2:
  client:  { ip: "0.0.0.0", port: "8080" }
  console: { ip: "0.0.0.0", port: "8082" }
```

### Endpoints (default)

Client (`/api/v1`):
- POST `/auth/login`               (identity facade)
- POST `/auth/logout`              (CheckLogin)
- GET  `/auth/me`                  (CheckLogin)
- GET  `/user/me`                  (CheckLogin)

Console (`/api/v1`):
- POST `/auth/login`               (identity facade)
- POST `/auth/logout`              (CheckLogin + admin)
- GET  `/auth/me`                  (CheckLogin + admin)
- GET  `/iam/subjects`             (CheckLogin + admin)
- GET  `/iam/subjects/:id`         (CheckLogin + admin)
- GET  `/users`                    (example management; CheckLogin + admin)
- GET  `/users/:id`                (example management; CheckLogin + admin)
```


