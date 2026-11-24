# lite-api

轻量自托管 AI API 网关。参考 New API 的多渠道聚合与 Sub2API 的账号池调度思路，面向标准 SDK 和 AI Agent 独立实现。

> 当前状态：**S0 工程初始化**。已有健康检查、前端工程状态页及 PostgreSQL/Redis 本地开发环境；模型转发、认证、调度和账务尚未实现，不能作为生产 AI 网关使用。

## 技术栈

- 后端：Go 1.27、Gin、net/http。
- 数据：PostgreSQL 17、pgx v5、sqlc、Goose SQL 迁移；持久化恢复任务使用 PostgreSQL。
- Redis：短期缓存、并发租约、限流及会话亲和。
- 前端：根目录 web/，Vue 3、TypeScript、Vite、Vue Router、Pinia、TanStack Vue Query、Tailwind CSS，按需引入 shadcn-vue、ECharts。
- 运维：Docker Compose、slog；生产构建计划将前端静态资源嵌入 Go 二进制，后续接入 Prometheus 和 OpenTelemetry。

业务依赖在对应开发阶段引入；当前 go.mod 与 web/package.json 只包含初始化工程实际所需的依赖。

## 目录

    cmd/lite-api/          Go 入口
    internal/config/       运行参数
    internal/httpserver/   HTTP 路由与测试
    web/                   唯一前端工程
    docs/                  需求、架构、API 契约、开发手册
    AGENTS.md              Agent 开发规则
    compose.yaml           开发用 PostgreSQL / Redis

后续数据库迁移、查询和生成代码分别位于 db/migrations/、db/queries/、internal/store/，按实际业务逐步建立。

## 本地启动

需要 Go 1.27、Node.js 24、npm 和 Docker Compose v2。

    docker compose up -d
    go mod tidy
    go run ./cmd/lite-api

在另一个终端：

    cd web
    npm install
    npm run dev

前端 http://127.0.0.1:5173，健康接口 http://127.0.0.1:8080/healthz。Vite 将 /api 请求代理至本地 Go 服务。

Go 服务读取 APP_ENV 和 HTTP_ADDR 环境变量，默认绑定 127.0.0.1:8080；不自动加载 .env 文件。.env.example 提供配置示例。S0 未接入数据库或 Redis 客户端，Compose 内的本地密码仅供开发。

## 验证

    go test ./...
    go vet ./...
    cd web && npm run typecheck && npm run build

## 路线与文档

S1：私有单渠道 MVP，包含真实请求、鉴权、基础调度及额度预留/结算。
S2：Responses、Anthropic、Gemini、完整账号池、管理控制台和生产部署。
S3：跨协议转换、授权凭据生命周期、监控及能力增强。商业化和规模化按需立项。

- [产品需求与验收](docs/requirements.md)
- [技术架构](docs/architecture.md)
- [API 契约](docs/api-contract.md)
- [开发指南](docs/development.md)
- [AGENTS.md](AGENTS.md)

## 许可证

MIT。架构和功能独立设计，仅接入获得授权且符合上游服务条款的服务。
