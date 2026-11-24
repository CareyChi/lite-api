# lite-api 技术架构

## 1. 总体架构

采用 Go 模块化单体，单一 HTTP 进程承载网关 API 和管理 API。前端独立工程位于仓库根目录 web/；开发时由 Vite 提供界面和 API 代理，生产部署方案在 S1 之后确定。

    Client / SDK / Agent
             |
      HTTP + Auth Boundary
             |
      Model / Policy Routing
             |
       Account Scheduler
             |
       Protocol Adapter
             |
       Upstream Provider
             |
        Usage Settlement
         /          \
    PostgreSQL      Redis

每个模块通过明确接口协作，不从 HTTP Handler 直接访问多个持久化模型并夹杂路由与计费逻辑。

## 2. 技术选型

| 领域 | 技术 | 约束 |
| --- | --- | --- |
| HTTP | Go 1.27 / Gin | SSE、Context 取消、共享连接池 |
| 领域持久化 | PostgreSQL / Ent / pgx | 用 pgx 处理必要的底层 SQL，Ent 管理领域实体 |
| 迁移 | Atlas | 审查式迁移，不静默修改生产 schema |
| 缓存与协调 | Redis | 限流、槽位租约、短期会话与健康状态 |
| 后台任务 | Redis Streams | consumer group、幂等消费、死信处理 |
| Web | Vue 3 / TS / Vite | Pinia 客户端状态；TanStack Query 服务端状态 |
| UI | Tailwind / shadcn-vue / ECharts | 按需引入，避免多套全局组件库 |
| 可观测性 | slog、OpenTelemetry、Prometheus | 不默认暴露用户内容与凭据 |

Ent、pgx、Atlas 和生产观测组件是目标技术栈；S0 不因尚无业务 schema 而提前生成空 ORM 代码。

## 3. 领域模块边界

- identity：后台用户、权限、客户端 API Key、后台会话。
- provider：渠道类型、账号凭据、能力标签、模型映射及加密配置。
- protocol：OpenAI、Anthropic、Gemini 各自的入站/出站解析与流式事件处理。
- routing：根据模型能力、健康、权重及策略选择候选池。
- scheduling：账号并发槽、租约、速率限制、冷却和粘性会话。
- billing：价格版本、额度预留、实际结算、不可重复写入的账本。
- usage：请求状态、用量来源、聚合统计与审计。
- jobs：异步消费者、对账、同步任务。
- httpserver：仅负责路由、错误映射、认证边界、请求上下文和响应传输。

建议每个模块拥有自己的 domain、service 和 repository 边界；实际创建目录和接口应由首个业务用例驱动，不预建一套空框架。

## 4. 协议适配策略

1. 相同协议默认尽可能使用原生透传路径，避免 JSON 反序列化/重组破坏供应商特定字段。
2. 跨协议转换采用显式能力映射；对 role、tool call、reasoning signature、缓存标记、图片/音频、结束原因和计费字段分别测试。
3. 不兼容的功能要明确失败，不能丢字段后返回表面成功。
4. 流式优先支持 SSE；后续如需 WebSocket/Realtime，单独建立传输层能力。
5. 使用同一个 request_id 关联调度、上游请求、用量与账本，不将完整请求体写入常规日志。

## 5. 调度与并发设计

- 以 route_key = 逻辑模型 + 协议能力 + 用户组确定候选上游。
- 路由过滤先于权重排序；未满足能力、停用、熔断、余额或并发约束的账号不可入选。
- Redis 使用原子脚本获取槽位/短期租约，按有效租期释放与回收，避免多节点超额并发。
- 健康信号区分认证失败、容量耗尽、上游 429、连接超时和业务 4xx；冷却策略按错误类型配置。
- 一旦已发送客户端首个响应字节，不再对生成请求做跨上游重试。

## 6. 计费与数据一致性

PostgreSQL 是唯一账本事实源。建议的核心表（S1/S3 随业务迭代实现）：

- users、api_keys、sessions、audit_events
- providers、provider_accounts、provider_models、model_routes
- requests、usage_records、price_versions
- wallets、reservations、ledger_entries
- jobs（可选任务持久化）与统计聚合表

账本条目不可就地覆盖历史事实；采用请求幂等键 + 唯一约束 + PostgreSQL 事务保证一次结算，预留和最终扣减可追溯。Redis 缓存可重建，不作为资产余额来源。

使用最小货币单位的整数或数据库 numeric 存储费用；不使用浮点数表示结算金额。缓存 Token 和特殊计费项必须独立记录。

## 7. 机密数据与隔离

- 客户端 API Key 保留不可逆校验摘要，创建时仅返回一次。
- 上游凭据加密后存入数据库，运行密钥通过部署环境注入，不写入仓库与日志。
- 校验自定义上游地址的 scheme、域名/IP、解析结果与重定向目标，阻止访问内网元数据地址。
- 管理 API 与模型 API 独立鉴权、授权与限流；会话 Cookie 使用安全属性。
- 对外代理头只信任明确配置的反向代理，不能直接信任客户端传入的 X-Forwarded-*。

## 8. 部署与可用性路线

S0：Go 开发服务 + Vite 前端 + PostgreSQL/Redis 本地 Compose。
S1：构建可部署的 API 服务与前端静态资源；版本化迁移与健康探针。
S2/S3：可靠的异步消费、备份、故障恢复与清晰的优雅停机顺序。
S4：根据实测负载决定扩容和 OpenTelemetry/Prometheus 接入时机，不预先拆微服务。
