# lite-api 技术架构

## 1. 架构原则

采用模块化单体。单个 Go 进程承载管理 API、模型 API、静态管理端和轻量后台任务；数据库和 Redis 作为独立基础设施。控制面和转发面有独立路由、中间件及授权边界，不拆成独立微服务。

    SDK / Agent                         Browser
        |                                  |
    Model API                         Admin API / Web
        |                                  |
    Authentication + Admission         Identity / Configuration
        |
    Routing + Scheduler
        |
    Gateway Orchestrator
        |
    Native Protocol Adapter + Credential Strategy
        |
    Shared HTTP Transport ----------> Upstream
        |
    Usage Observation + Settlement
        |
    PostgreSQL (truth / jobs)      Redis (short-lived coordination)

Go context 用于取消 HTTP 请求和流式传输；账户租约、用量及账务完成由独立、有界的收尾过程保障，不能跟随客户端 context 被直接丢弃。

## 2. 技术栈

| 层级 | 选型 | 约束 |
| --- | --- | --- |
| 运行时 | Go 1.27、Gin、net/http | Gin 负责 HTTP 边界，网关传输与协议解析不依赖 Gin Context |
| 持久化 | PostgreSQL 17、pgx/v5、sqlc | 手写 SQL，生成类型安全查询，单一数据库访问路径 |
| 迁移 | Goose（SQL） | 版本化迁移、显式执行、工具版本固定 |
| 缓存与调度 | Redis 7、go-redis/v9 | Lua 原子限流、租约及短期状态；无资产事实 |
| 持久任务 | PostgreSQL | 任务表、租约、FOR UPDATE SKIP LOCKED；跨模块事件需要时使用 Transactional Outbox |
| 前端 | Vue 3、TypeScript、Vite、Vue Router | 唯一前端工程位于根目录 web/ |
| 状态与界面 | Pinia、TanStack Vue Query、Tailwind CSS、shadcn-vue、ECharts | 服务端状态和客户端 UI 状态分离，组件与图表按需引入 |
| 管理 API | OpenAPI 3.1 | 实现时维护契约，生成 TS 类型；原生模型协议独立 |
| 可观测 | slog、Prometheus、OpenTelemetry | 先结构化日志与关键指标，按需部署 Grafana/追踪收集器 |
| 交付 | Go embed、Docker Compose、GitHub Actions | 单一发布二进制，Linux AMD64/ARM64，构建与迁移可重复 |

不使用 Ent/GORM、Atlas、Redis Streams、独立 MQ 或微服务作为首版基础依赖。sqlc、Goose 与其他业务依赖在首次使用时引入并固定版本，禁止手造未验证的依赖锁文件。

参考：[sqlc/pgx](https://docs.sqlc.dev/en/latest/guides/using-go-and-pgx.html)、[Goose SQL migrations](https://pressly.github.io/goose/documentation/annotations/)、[PostgreSQL 锁定语句](https://www.postgresql.org/docs/17/sql-select.html)。

## 3. 目录与模块

实际创建目录由业务用例驱动，不预建空接口或空 package。

    cmd/lite-api/         进程入口、命令
    internal/config/      类型化运行配置
    internal/httpserver/  路由、认证边界、管理 API、Web
    internal/identity/    用户、密钥、管理会话、权限
    internal/provider/    渠道、上游账号、凭据、模型与能力
    internal/gateway/     请求执行、上游尝试、生命周期编排
    internal/protocol/    原生协议解析、转发和可选的协议转换
    internal/routing/     策略过滤与候选选择
    internal/scheduling/  限流、并发租约、冷却、短期亲和
    internal/transport/   上游 HTTP、连接池、流式 I/O、地址策略
    internal/billing/     价格、额度预留、结算、账本
    internal/usage/       请求与尝试记录、计量、聚合
    internal/jobs/        持久任务领取、恢复与对账
    internal/store/       pgxpool、事务边界、sqlc 生成代码
    db/migrations/        Goose SQL 迁移（唯一 schema 来源）
    db/queries/           sqlc SQL 查询
    web/                  Vue 独立工程、生产静态资源
    docs/                 需求及接口契约

由 gateway 编排身份、路由、账务和上游执行；适配器不写账本、不直接选择账号；调度不解析供应商请求体；HTTP Handler 不持有跨模块业务事务。领域之间以具体用例所需的窄接口协作。SQL 生成类型只在持久层及明确的映射边界使用，不直接作为公开 API DTO。

## 4. 协议、Provider 与凭据

### 4.1 原生优先

OpenAI Chat Completions、OpenAI Responses、Anthropic Messages、Gemini 分别识别为协议/端点族；供应商品牌不是协议类型。

每个适配器负责请求验证、提取路由与预算所需的最小元数据、构建受控上游请求、解析原生错误、观察用量与必要的事件信息。原生转发保留其他合法 JSON 字段，尽量不重组完整数据结构；仅修改路由、凭据和明确政策必需的字段。

统一的内部对象描述模型、所需能力、请求预算、用量观测和状态，而不是把所有请求强制转换成 OpenAI Chat Completions 格式。跨协议转换按方向建立能力矩阵；签名、推理块、工具调用、媒体、缓存控制及状态语义不可映射时应拒绝。

### 4.2 状态型请求

区分两类亲和：

- 软亲和：在授权用户/API Key、逻辑模型、路由和显式会话标识作用域内优先选择原账号，优化缓存；账号失效时可重新选择。
- 硬绑定：对 previous_response_id、conversation 等引用上游状态的请求，依据已持久化的所有权、上游账号/状态域和期限恢复路由。不能因为账号冷却就将原 ID 发往不同状态域。找不到映射或没有权限时返回明确错误。

持久化绑定只保存必要引用、作用域与有效期，不默认存储完整对话。不得静默修改 store 等影响隐私或上游状态保留的参数。仅支持已实现并验证的状态生命周期操作。

### 4.3 凭据策略

渠道负责协议及端点；账号负责凭据、权限、资源预算及健康；凭据策略独立构建上游认证。S1 支持 API Key；后续经授权的 OAuth 等策略可提供刷新及过期处理，通过锁/版本检查避免多个实例覆盖新凭据。不得在 Provider 中复制整个请求编排或伪造未授权的客户端身份。

## 5. 传输、重试和 SSE

- HTTP Transport/Client 按兼容的网络策略复用，连接池有界；分别控制连接、TLS、响应头等待、流式空闲和最大持续时间。不得设置会无条件截断所有长流的全局写超时或 Client.Timeout。
- 设定最大入站请求体和可处理的单个 SSE 事件；原生 SSE 增量转发、及时 flush、按帧观察用量，不累积全部文本。适配器必须保留事件序列和合法扩展，不用简单逐行假设替代正确的 SSE 解析。
- 下游慢读时施加有界背压；取消或超时关闭上游响应体、释放资源。反代必须关闭流式缓冲，避免修改 SSE 分帧语义。
- 区分内部 request_id、外部幂等标识与 attempt_id；每次向上游发送前持久化尝试状态。不能仅因下游还没收到首个字节就认为请求可以重发。
- 仅在能够确认未提交上游、上游明确拒绝且语义允许重试，或具有经验证的上游幂等机制时，才允许有限重试。可能被接受且结果未知时记录待对账，不自动重发；提交下游响应头后停止故障切换。
- 注意 HTTP Transport 的隐式重放行为；不盲目传递客户端认证、Cookie、代理头或幂等头。实际重试由明确的策略和测试验证。
- 流开始后的错误按对应协议允许的错误事件处理，否则终止连接；不能补写第二份 HTTP 状态或伪造结束事件。

## 6. 路由与 Redis 协调

1. 根据用户授权、逻辑模型、端点、请求能力和状态绑定过滤候选。
2. 按优先级分层，在当前层对健康账号进行加权选择；拒绝的候选不消耗上游生成尝试。
3. 对选择结果原子领取相关并发槽和速率额度；失败时尝试仍合格的其他候选。多个约束在同一个 Redis 原子操作中处理或定义可靠的补偿，不使用本地读取后更新计数。
4. 使用唯一租约标识、到期时间与条件续租/释放；不能由旧租约释放新持有者的槽位。续租失败时停止继续使用失去租约的上游资源，并记录不确定成本。
5. 认证失败禁用或告警；429/容量失败按上游信号设置冷却；客户端参数错误不降低账号健康；超时分类后处理，避免雪崩。
6. Redis 不可用、重启丢失活跃租约或失去一致状态时，按明确的恢复门槛停止新的受限请求，不能假设旧连接均已退出。

单机 Redis 为首版部署形态；不承诺 Redis 故障转移下严格全局并发上限。多节点上线前验证相应故障域、租约重建和降级策略。Redis Key 均带前缀、作用域及合理 TTL，缓存不是权威权限或余额来源。

## 7. 数据与计费

### 7.1 数据模型

S1 按垂直用例实现：

- users、api_keys、sessions、audit_events
- providers、provider_accounts、models、model_routes
- requests、request_attempts、usage_records
- price_versions、wallets、reservations、ledger_entries
- jobs（持久化恢复任务）

S2 增加必要的 protocol_state_bindings 和统计聚合；后续只有需要可靠发布业务事件时才增加 outbox_events。

采用外键、非空、唯一索引、检查约束保证核心不变量。时刻使用 UTC timestamptz；数据模型不将上游密钥、用户身份和请求标识混为一体。

### 7.2 SQL 与事务

Goose 版本化 SQL 迁移为 schema 唯一事实来源；sqlc 从迁移和 db/queries/ 生成 pgx v5 查询代码。使用 pgxpool 和明确事务边界，将 sqlc Queries.WithTx 绑定同一事务。账本、余额、任务创建必须原子提交。关键事务用行锁或条件 UPDATE，并在稳定锁顺序下执行；并发冲突按明确 SQLSTATE/事务语义有限重试。不得在调用外部 HTTP 的整个持续时间持有事务。

### 7.3 请求与尝试状态

逻辑请求有唯一 request_id；一个请求可存在多个 attempt_id。记录预留、请求状态、各次尝试状态和用量来源。

    admitted -> reserved -> in_flight -> settled
                                     -> pending_reconciliation -> settled / adjusted
                         -> rejected_or_cancelled_before_send -> released

上述为逻辑生命周期，不要求提前生成完整状态机框架。发送前持久化尝试；进程在发送与记录结果之间崩溃时，恢复过程不能判断上游一定未执行。已提交结算任务只重复执行幂等数据库操作，不重放生成请求。

### 7.4 金额和结算

金额和价格使用固定精度 numeric（初始规划 numeric(38,12)，按实际单位及上界校验），Go 中采用精确十进制或 pgx numeric 类型；禁止 float32/float64 作为财务数据。使用量以整数及计费维度存储；价格版本包含币种/额度单位、计价基数、计量规则和舍入策略。请求固定售卖价格；每次尝试记录对应上游价格。

钱包维护可用额与预留额的事务化快照，流水只追加；预留、扣减、释放和人工调整各自具有唯一业务幂等键。每个操作的余额变化与流水在同一事务；不静默透支、不修改历史流水。上游成本与用户收费分别记录，不能把失败尝试成本自动重复转嫁给用户。

预算必须约束已支持的输入/输出和可计费工具调用；如果实际费用仍超预留，按明确的结算政策处理差额并留下审计，不以负数余额掩盖。未回报用量应标记 unknown/estimated/reported 等来源，区分取消、失败和待对账。用户断线不能直接释放可能已产生费用的预留。

结算独立于客户端 Context，设置有限超时。若数据库写入失败，依据已有请求、预留和尝试恢复；保留未知状态直至对账期限，届时按明确业务政策释放未确认预留或结算经确认的部分，并保留未知上游成本供后续调整。不能承诺对第三方调用恰好执行一次。

## 8. 持久化后台任务

首版用 PostgreSQL 任务表处理用量恢复、对账和后续统计等低至中频任务，不引入额外消息中间件。

- 在同一业务事务中插入必要的任务或 outbox 事件。
- worker 使用短事务、FOR UPDATE SKIP LOCKED 领取有限数量的任务，持久化 lease_owner/lease_until 后提交；在事务外执行任务。
- 按任务类型及幂等键消费，条件更新租约状态，使用有限退避、最大尝试次数、死信状态和人工重试；任务重复领取不应重复扣款。
- 有界轮询（空闲退避、可选 LISTEN/NOTIFY 唤醒），通知只用于唤醒，不替代任务持久化。
- 只在实际需要可靠事件发布时使用 Transactional Outbox，禁止在数据库提交与非事务消息发布之间形成不可恢复的双写。

## 9. 安全

- 网关 Key 为高熵随机值，存储不可逆摘要与可识别前缀；管理密码使用 Argon2id。会话由服务器管理并支持吊销；Cookie 具 HttpOnly、Secure（HTTPS）及适当 SameSite，状态修改需要 CSRF 防护。
- 上游凭据使用 AEAD（例如 AES-GCM）加密，版本化主密钥由部署环境或 secret 文件注入，支持计划内轮换；不与数据库备份一起公开保存。
- 公共上游默认 HTTPS；管理员显式授权的私有上游通过受控地址/网段策略接入。禁止访问云元数据、网关控制面和未授权内网/回环；在实际建立连接时验证解析地址，防 DNS rebinding；重定向默认禁止，授权时逐跳验证且不跨信任域带认证头。
- 可信反代显式配置；不信任任意客户端 X-Forwarded-*。外部代理作为独立信任边界，不默认继承环境 HTTP_PROXY；需要代理时配置受控出口。
- 使用同源管理端；API 层权限、数据隔离、请求限额、日志脱敏不能依赖前端。避免记录认证查询参数及可能携带敏感内容的上游错误。
- 明确请求/日志保留周期，统计不能暴露其他用户或上游账号的秘密。

## 10. 部署与可观测性

S0：Go HTTP 骨架、Vite 开发服务、Compose 数据依赖。

S1/S2：生产构建先编译 web/，通过 Go embed 嵌入静态产物（嵌入入口放在 web/ 中，使 go:embed 不跨目录），或等价的构建期复制方案。开发后端不依赖前端 dist 存在。管理端路由归 /、/console 或明确前缀，模型 API 和 /api/v1/ 不进入 SPA fallback。迁移独立命令执行；Go 程序以非 root 运行，外部反代终止 TLS。

/healthz 只表示进程存活；业务接入后新增 /readyz 检查数据库、Redis 和 schema 状态。停机先停止新请求、关闭就绪状态，给正在进行的流式请求有界排空时间，再取消传输并完成可恢复的结算任务。备份恢复需覆盖数据库和凭据密钥，Redis 重启有调度恢复策略。

记录请求总耗时、首事件/首 Token（定义可观测范围）、上游尝试、错误率、流式并发、待结算任务和未知用量等关键指标；Prometheus 标签不使用 request_id、user_id 或任意模型别名等无界高基数值。Grafana/追踪收集器非必选运行依赖。
