# lite-api

基于 Go 与 Vue 3 的自托管 AI API 网关。架构设计参考 New API 的多渠道聚合能力与 Sub2API 的账号池调度思路，代码独立实现。

> 当前状态：**工程初始化阶段**。已提供后端健康检查、前端状态页面和本地 PostgreSQL/Redis 开发环境；尚未实现上游请求转发、身份认证、账号池、计费或管理功能。不要将当前版本投入公网运营。

## 技术栈

- 后端：Go 1.27、Gin；后续按需求引入 Ent、pgx、Atlas。
- 数据与异步：PostgreSQL、Redis、Redis Streams。
- 前端（仓库根目录 web/）：Vue 3、TypeScript、Vite、Pinia、TanStack Vue Query、Tailwind CSS；业务组件采用 shadcn-vue，统计图表采用 Apache ECharts。
- 运维：Docker Compose、结构化日志；后续接入 OpenTelemetry、Prometheus 与 Grafana。

## 仓库结构

    cmd/lite-api/          Go 服务入口
    internal/config/       运行参数
    internal/httpserver/   HTTP 路由与测试
    web/                   独立的前端工程
    docs/requirements.md   产品需求与验收条件
    docs/architecture.md   模块边界和设计约束
    docs/api-contract.md   对外协议与接口约定
    docs/development.md    本地开发与验证
    AGENTS.md              编码 Agent 工作规则
    compose.yaml           本地 PostgreSQL 与 Redis

## 本地启动

需要 Go 1.27+、Node.js 24+、npm 和 Docker Compose。

1. 启动依赖服务：

       cp .env.example .env
       docker compose up -d

2. 启动后端（新的终端）：

       go mod tidy
       go run ./cmd/lite-api

3. 启动前端（新的终端）：

       cd web
       npm install
       npm run dev

访问 http://localhost:5173。后端健康检查为 http://localhost:8080/healthz；前端使用 Vite 代理访问 /api/v1/system/info。

当前后端只读取 APP_ENV 和 HTTP_ADDR；数据库连接、Redis 客户端与迁移将在后续阶段实现。开发环境中的数据库密码仅可用于本地测试。

## 验证

    go test ./...
    go vet ./...
    cd web && npm run typecheck && npm run build

完整开发规范见 docs/development.md。

## 设计原则

1. 原生协议优先透传，跨协议转换按能力明确声明，不假装无损转换。
2. 请求转发、账号调度、鉴权和计费相互解耦，保持模块化单体。
3. PostgreSQL 是计费账本的最终事实源；Redis 只承担缓存与并发协调。
4. 上游订阅或账号接入仅限获得授权且符合对应服务条款的方式。
5. 保留仓库原有 MIT LICENSE；不直接复制 New API 或 Sub2API 的受许可证约束代码。

## 需求与实施

- [产品需求](docs/requirements.md)
- [技术架构](docs/architecture.md)
- [协议及接口](docs/api-contract.md)
- [开发手册](docs/development.md)
