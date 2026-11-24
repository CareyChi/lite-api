# 开发与验证指南

## 1. 环境

- Go 1.27 或更新的兼容稳定版。
- Node.js 24 LTS、npm。
- Docker Engine / Docker Compose v2。
- 推荐使用 Linux/macOS；Windows 可用 WSL2。

## 2. 首次启动

在仓库根目录执行：

    cp .env.example .env
    docker compose up -d
    go mod tidy
    go run ./cmd/lite-api

默认监听 http://127.0.0.1:8080（后端绑定地址为 :8080）。另开终端：

    cd web
    npm install
    npm run dev

前端开发地址为 http://localhost:5173，/api 路径由 Vite 代理到 http://127.0.0.1:8080。

.env 是本地设置示例，目前后端只读取 HTTP_ADDR 与 APP_ENV；数据库和 Redis 连接尚未接入 Go 服务。PostgreSQL/Redis Compose 配置只用于开发，不构成生产部署方案。

## 3. 开发约定

- API 业务逻辑放 internal/，cmd/ 只负责进程入口。
- web/ 为唯一前端工作区，不在其他目录另建前端工程。
- 前端服务端状态使用 TanStack Query，客户端 UI 状态放 Pinia。
- 业务 schema 到位后使用 Ent 和 Atlas，迁移脚本必须入库。
- 跨协议适配与 SSE 需要以真实的事件序列设计测试 fixture，不能仅做 JSON 快照。
- 管理端页面必须绑定真实服务端能力；当前仅显示工程状态。

## 4. 本地自检

后端：

    gofmt -w cmd internal
    go test ./...
    go vet ./...

前端：

    cd web
    npm run typecheck
    npm run build

若改变 API 契约、数据库或产品需求，先同步更新相关 docs 内容。新增依赖需将 Go go.sum 与前端 lockfile 纳入版本管理。初始化阶段因无可用外部依赖下载环境，仓库尚未附带生成的锁文件；在首次可联网安装后应提交 go.sum 与 web/package-lock.json，并在 CI 改用锁文件安装。

## 5. 集成测试与发布

- 先用可控的模拟上游测试超时、断流、429、失败重试及并发租约回收。
- 账务场景必须覆盖重复请求、重复消费者、断线和上游未报告用量。
- 运行本地验证后再使用 CI 作最终验收，避免频繁触发远程构建。
- 未被明确要求发布版本时，仅提交代码，不创建 Release/Tag。
- 不上传真实 API Key、数据库备份、Token、环境密钥和客户请求正文。
