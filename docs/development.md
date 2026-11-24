# 开发与验证指南

## 1. 工具链与启动

- Go 1.27 的最新稳定补丁版本；Node.js 24 LTS、npm；Docker Compose v2。
- 推荐 Linux、macOS 或 WSL2。

仓库根目录：

    docker compose up -d
    go mod tidy
    go run ./cmd/lite-api

另开终端：

    cd web
    npm install
    npm run dev

默认 Go 服务监听 127.0.0.1:8080，Vite 监听 127.0.0.1:5173，代理 /api 到 Go。

.env.example 为环境变量示例。当前 Go 进程只读取 APP_ENV、HTTP_ADDR，不自动读取 .env；需要自定义时通过 shell 或部署工具显式注入。例如 HTTP_ADDR=127.0.0.1:18080 go run ./cmd/lite-api；自定义后端端口后同步调整 Vite 代理。DATABASE_URL、REDIS_ADDR 待 S1 接入。开发 Compose 不构成生产部署方案。

## 2. 数据访问与生成

S1 随首次业务 schema 引入 pgxpool、sqlc 和 Goose，不预建空表。

- db/migrations/：按版本排序的 Goose SQL 迁移；已发布的迁移不得改写，新增迁移要验证空库安装与旧库升级。
- db/queries/：sqlc 查询；从迁移生成，sqlc.yaml 与工具固定版本入库，生成代码放 internal/store/ 的生成包。
- 显式执行迁移；应用启动检查 schema 兼容性，不自动执行破坏性变更。不可逆迁移需要备份和恢复说明。
- 业务事务使用 pgx 与 sqlc 的 WithTx；不在请求上游期间占用数据库事务。
- 代码生成需要可重复执行；CI 检查 sqlc 生成结果与仓库一致。只修改原 SQL，不能手改生成代码。

## 3. 开发顺序

以需求文档 S1–S4 为基准；每次交付一个可验证的垂直用例。S1 首先打通用户/API Key、模型路由、上游调用、用量和计费账本闭环，再扩展 S2 原生协议、多账号与管理端。

管理 API 同步 OpenAPI 契约与前端类型；真实客户端协议通过独立 fixture 验证。使用真实 PostgreSQL/Redis 集成测试验证锁、租约和事务，不用内存模拟宣称数据库一致性通过。

## 4. 本地自检

后端：

    gofmt -w cmd internal
    go test ./...
    go vet ./...

涉及并发的模块增加 go test -race；涉及数据层的模块按实际测试配置启动 PostgreSQL/Redis 执行集成测试。前端：

    cd web
    npm run typecheck
    npm run build

重点回归：SSE 分帧、取消、慢消费者、工具调用及未知用量；并发预留、重复结算、进程中断和 Redis 丢失租约；可重试拒绝与上游结果未知必须区分。

## 5. 依赖与 CI

S0 尚无 go.sum 与 web/package-lock.json。必须在有完整 Go/Node 工具链及联网依赖安装的环境执行 go mod tidy、npm install，提交真实生成的锁文件并验证构建，才算完成 S0 验收。届时 CI 使用 go mod download、go test、go vet 和 npm ci，不再在验证过程中动态生成依赖锁文件；禁止手工编造校验和或 lockfile。

当前 CI 用于初次验证。以后固定 Go 模块和 npm 锁定依赖，固定生成工具版本及构建环境；CI 不承担日常调试，不高频轮询。生产构建先构建 web 静态资源，再嵌入 Go 二进制，开发构建不依赖已有 dist。

## 6. 部署与提交

- S2 交付单机 Docker Compose、非 root 服务、健康与就绪探针、备份恢复、升级及流式优雅停机。后续按负载决定扩容。
- 凭据、主密钥、数据库备份、客户请求正文和真实 API Key 不入库；使用授权的上游，仅在必要时进行付费冒烟测试。
- 先执行本地检查再提交；无法执行的检查明确说明。未被要求发布时不创建 Tag/Release 或更改产品版本号。
