# lite-api · AGENTS.md

适用于仓库全部目录。执行任务前阅读本文件及相关需求和架构文档。

## 工程约束

- 项目为 lite-api，Go 模块路径 github.com/CareyChi/lite-api。保留 MIT LICENSE，独立实现，不直接复制 New API / Sub2API 的受许可证约束代码。
- 唯一前端工程位于根目录 web/，Go 入口为 cmd/lite-api，领域代码置于 internal/。
- 技术栈：Go 1.27 + Gin/net/http + PostgreSQL 17 + pgx v5 + sqlc + Goose + Redis；Vue 3 + TypeScript + Vite + Vue Router + Pinia + TanStack Vue Query + Tailwind CSS + shadcn-vue。按实际用例引入依赖。
- 模块化单体；不引入 Ent/GORM、Atlas、Redis Streams、独立 MQ 或微服务，不提前创建空模块、接口和占位业务。
- 文档中的计划不等于已实现功能。当前 S0 仅为工程骨架。

## 文档

- docs/requirements.md：产品范围、交付阶段与验收条件。
- docs/architecture.md：模块边界、数据、路由、协议与安全设计。
- docs/api-contract.md：原生端点、错误、状态绑定和接口契约。
- docs/development.md：开发、构建、迁移与验证规范。

仅当需求、契约或技术设计变化时更新对应内容，不把任务描述、临时方案或修改记录当作代码注释和文档更新日志。

## 后端

- HTTP Handler 负责边界；gateway 编排请求，protocol 管理协议，provider 管理凭据，routing/scheduling 管理候选和租约，billing 管理计费。不允许业务模块绕过明确边界。
- 原生协议优先，不强制转换成统一 Chat Completions JSON。合法扩展字段、工具调用、思考签名、缓存和状态语义不得静默丢失；不兼容能力明确拒绝。
- 复用标准 HTTP Transport；SSE 增量转发、正确解析事件、及时 flush、支持取消和有界背压；不能设置会截断正常长流的全局写超时。
- 每次上游尝试独立记录。只有上游明确未接受、可重试拒绝或有已验证幂等机制且下游未提交响应时才能重发。结果未知不自动重试。
- 状态型 ID 必须验证所有权和上游状态域。软粘性仅作缓存优化，不替代硬绑定及权限检查。
- PostgreSQL 是账务事实源：用 pgx + sqlc、显式事务、唯一约束、精确数值和不可变流水实现预留及结算；不使用浮点金额或 Redis 余额作最终依据。
- 不持有跨上游 HTTP 调用的数据库事务。结算与资源释放必须独立于客户端取消，失败可根据持久化请求/尝试恢复，不能用重发生成代替对账。
- Redis 的限流和槽位租约须原子、可续期、按所有者条件释放；Redis 丢失状态或不可用时不绕过受限付费请求。
- PostgreSQL 承载持久任务，领取租约、幂等消费、有限退避及失败状态；只有实际存在可靠事件发布需求才实现 Transactional Outbox。
- db/migrations/ 是 schema 唯一来源；db/queries/ 保存 sqlc SQL，生成代码纳入版本管理，不手改生成文件。迁移独立执行，不隐式破坏生产数据库。
- 上游凭据加密、网关密钥不可逆摘要、后台密码安全哈希；日志不能暴露凭据或完整请求/响应正文。管理授权、SSRF、实际 DNS 解析地址验证及可信反代是上线门槛。
- 只接入获得授权的上游认证方式，不伪造未授权客户端身份或规避服务限制。

## 前端

- 所有前端源码位于 web/；Composition API + TypeScript，Vue Router 管理页面，Pinia 管理客户端状态，TanStack Query 管理服务端数据。
- API 层统一调用管理接口；依据真实响应渲染加载、错误、空状态和权限，金额字符串不能转成浮点进行账务计算。
- Tailwind CSS + 按需 shadcn-vue；统计采用 ECharts，不混入第二套全局 UI 框架。
- 未实现能力不显示为已可用，不制造假统计或无效操作。生产静态产物通过构建嵌入 Go 服务，开发模式不要求 dist。

## 开发与验收

- 根据当前交付阶段按垂直用例完成后端、迁移、最小界面与测试，优先实现 S1 的完整请求和计费闭环，再扩展协议与账号池。
- Go：gofmt、go test ./...、go vet ./...；并发、租约、账务变更增加真实 PostgreSQL/Redis 集成测试及适用的 race 测试。
- 前端：在 web/ 执行 npm run typecheck、npm run build；业务交互复杂后增加针对性测试。
- 协议测试使用确定性的模拟上游；覆盖流式、断线、未知用量、明确拒绝与不确定提交。账务覆盖并发预留、重复消费、进程恢复，验证不会重复扣款或隐式透支。
- 新增/修改依赖时提交 go.mod、go.sum、web/package-lock.json 中实际生成的对应变更。固定生成及迁移工具版本，验证生成代码无差异。
- 优先本地自检，再提交并使用 CI 做最终验收，不高频轮询或以 CI 代替开发调试。无法执行的检查明确说明，不宣称通过。
- 除明确要求发布外，只提交代码，不修改版本号、不创建 Tag 或 Release。避免无关重构、重复注释和推测性兼容分支。
