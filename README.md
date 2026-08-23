# Salsa Circle

Salsa Circle 是一个完全离线可运行的莎莎舞会聚合与舞伴邀约演示平台。它把公开资料、舞会时段、解释型匹配、单次邀约、容量候补和履约信誉连接成一个可审计的业务闭环。

## 架构与目录

- `cmd/server`：进程入口、优雅停机和超时配置。
- `internal/domain`：用户隐私、舞会状态机、匹配策略、邀约、报名、审核和排行等领域模型。
- `internal/application`：用例与调用方接口，事务边界通过构造函数注入。
- `internal/repository/memory`：确定性本地演示适配器；`internal/repository/postgres` 和 `migrations` 提供 pgx/PostgreSQL 持久化边界。
- `internal/transport/http`、`internal/middleware`：Gin API、request_id、统一错误、CORS 和安全响应头。
- `internal/platform`：可替换时钟、幂等键、上传校验和 outbox 重试。
- `web`：Vue 3、TypeScript、Vite、Pinia、Element Plus 前端。

## 本地启动

需要 Go 1.24+；前端需要 Node 20+。后端默认监听 `:8080`，演示不依赖网络服务。

```bash
go run ./cmd/server
npm install
npm run dev
```

访问 `http://localhost:5173/events`。健康检查为 `/healthz`，就绪检查为 `/readyz`。HTTP API 统一前缀为 `/api/v1`，可查看 `api/openapi/openapi.yaml`。

## 容器与数据库

```bash
docker compose up --build
psql "$DATABASE_URL" -f migrations/001_init.sql
psql "$DATABASE_URL" -f migrations/002_seed.sql
```

Compose 同时启动应用和 PostgreSQL；应用镜像以非 root 用户运行。Dockerfile 使用官方 Go 多架构基础镜像，`TARGETARCH` 可用于 amd64/arm64 构建。

## 核心状态机

舞会依次经过 `draft -> review -> published -> completed`，审核前可以退回，已发布舞会可取消；非法跳转返回稳定错误码。邀约只能从 `pending` 处理一次并带 UTC 过期时间。报名在容量用尽时进入候补，取消在同一应用事务中释放名额并晋升下一位。

## 隐私与安全

匹配只读取明确公开字段。联系方式和精确坐标只有双方确认后才进入公开视图；日志不写入令牌、密码或附件正文。RBAC/资源归属由用例层校验，审计事件记录操作者、来源、前后差异、原因和 UTC 时间。

## 测试与限制

```bash
gofmt -w .
go test ./...
go test -race ./...
go vet ./...
npm run test
npm run typecheck
npm run build
```

仓储默认使用内存适配器以便离线演示；生产部署应配置 PostgreSQL、短期访问令牌密钥和可撤销刷新令牌存储。地图使用本地点位字段，不连接在线地图服务。
