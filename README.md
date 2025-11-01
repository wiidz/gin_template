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

方式一：一键脚本（推荐）

```bash
# 从本地模板复制（无网络）
scripts/new_project.sh \
  --module github.com/you/my_service \
  --template "/absolute/path/to/gin_template" \
  --dir "/absolute/path/to/my_service"

# 或直接远程克隆模板并初始化
scripts/new_project.sh \
  --module github.com/you/my_service \
  --dir "/absolute/path/to/my_service"

# 已在模板目录中，原地去模板化
cd "/absolute/path/to/gin_template"
scripts/new_project.sh --module github.com/you/my_service --in-place
```

方式二：使用 Go CLI（可编译为二进制，更适合全局调用）

```bash
# 推荐把所有 flags 放在项目名之前（Go flag 在遇到第一个非 flag 参数后停止解析）
go run ./cmd/gt new \
  --module github.com/you/my_service \
  --template "/absolute/path/to/gin_template" \
  --dir "/absolute/path/to/my_service" \
  my_service

# 或者本地编译后放到 PATH 中
go build -o ~/bin/gt ./cmd/gt && \
gt new --module github.com/you/my_service --dir "/absolute/path/to/my_service" my_service

# 或直接安装远程版本
go install github.com/wiidz/gin_template/cmd/gt@latest
```

- CLI 会自动探测模板目录；也可通过环境变量 `GIN_TEMPLATE_ROOT=/absolute/path/to/gin_template` 或参数 `--template` 指定。
- 可选参数包括 `--dir`（目标目录，必须不存在）、`--skip-git`、`--skip-tidy`。
- 未显式指定时，CLI 会自动将模板仓库克隆/更新到用户缓存目录（默认地址 `https://github.com/wiidz/gin_template.git`），无需手动拷贝模板。

方式三：手动克隆并去模板化

```bash
git clone https://github.com/wiidz/gin_template.git my_service
cd my_service
rm -rf .git cmd/gt tmp scripts/new_project.sh

# 修改 module 与导入路径
sed -i '' "s|^module github.com/wiidz/gin_template|module github.com/you/my_service|" go.mod
find . -type f \( -name '*.go' -o -name 'go.mod' -o -name '*.sum' -o -name '*.yaml' -o -name '*.toml' -o -name 'Makefile' \) \
  -not -path './.git/*' -not -path './cmd/gt/*' -not -path './tmp/*' \
  -exec sed -i '' "s|github.com/wiidz/gin_template|github.com/you/my_service|g" {} +

git init && git add . && git commit -m "bootstrap from template"
go mod tidy
```

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


