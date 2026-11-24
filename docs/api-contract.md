# API 契约与协议兼容要求

## 1. 当前可用接口（S0）

| 方法 | 路径 | 响应 |
| --- | --- | --- |
| GET | /healthz | 进程存活，JSON status=ok |
| GET | /api/v1/system/info | name、status=scaffold、api_version=v1 |

当前不存在身份认证、模型转发或账务。/healthz 不表示数据库、Redis 或网关业务已就绪。S1 接入 /readyz，生产环境不向不可信客户端暴露内部依赖细节。

## 2. 计划的模型 API

| 协议 | 方法及路径 | 阶段 |
| --- | --- | --- |
| OpenAI | GET /v1/models | S1 |
| OpenAI | POST /v1/chat/completions | S1 |
| OpenAI | POST /v1/responses | S2 |
| Anthropic | POST /v1/messages | S2 |
| Anthropic | POST /v1/messages/count_tokens | S2，按能力 |
| Gemini | POST /v1beta/models/{model}:generateContent | S2 |
| Gemini | POST /v1beta/models/{model}:streamGenerateContent | S2 |
| Gemini | POST /v1beta/models/{model}:countTokens | S2，按能力 |
| Gemini | GET /v1beta/models | S2，按能力 |

所有计划接口必须通过实现和兼容性测试才可宣称支持。Gemini 流式入口按已支持的查询参数（例如 alt=sse）处理；不同协议的模型列表保持其原生格式。

Responses 的读取、删除、取消、Conversation、后台模式和 WebSocket 等不是创建接口天然支持的能力。对于需要上游保存状态的操作，必须先定义状态绑定、所有权与生命周期。未实现的功能返回明确错误，不转发至任意账号或静默丢弃字段。

## 3. 鉴权与路由

- OpenAI 支持 Authorization: Bearer；Anthropic、Gemini 支持各自已实现的原生密钥传递方式，统一认证为网关 API Key。不同认证来源同时出现且值冲突时拒绝请求。
- 上游凭据只由服务端凭据策略注入；客户端 Authorization、x-api-key、x-goog-api-key、key 查询参数、Cookie 和内部代理头不作为上游凭据透传。
- 模型列表和请求路由必须遵守用户、Key、模型能力及状态绑定权限。不得让客户端指定未授权 Base URL 或内部账号。
- 模型映射只允许管理员定义的别名与路由；不静默替换为其他逻辑模型。
- request_id 由网关生成，响应头可返回该 ID；每次上游尝试内部记录独立 attempt_id。客户端提交的幂等标识不是网关请求 ID，不可据此假定生成请求可重放。

## 4. 请求、响应和流式语义

- 同协议原生优先；保留已支持的合法扩展字段、状态码、响应媒体类型、工具调用 ID/参数和事件顺序。仅替换路由、认证和明确政策所需字段，剥离 hop-by-hop、敏感及无权透传的头。
- 对上游状态型操作持久化用户/API Key 作用域、上游状态 ID、账号或已验证的状态域、有效期。后续请求必须命中同一状态域；找不到映射、权限不足、状态过期或账号不可用时返回错误。不能以软粘性会话代替硬绑定。
- 不默认修改 store、缓存控制、思考配置、工具选择或最大输出等影响语义的参数。预算需要约束时使用已公开的模型策略；超出限制可拒绝而不是隐式降低能力。
- SSE 按协议解析完整事件和必要用量，不把网络 chunk 当成事件，不整段缓存；及时 flush，保证事件顺序和终止语义。下游取消后关闭上游响应并启动独立的结算/租约收尾。
- 发送响应头之前且能确认上游未接受、明确可重试拒绝或有已验证的幂等机制时，才允许有限换路。上游已可能执行但结果未知时，不自动重发。
- 跨协议转换属于 S3：逐方向定义输入、输出、流事件、错误、工具调用、推理签名、缓存、媒体与状态保真矩阵。不能转换的能力返回不支持，不声称完全兼容。

参考：[OpenAI Responses 状态管理](https://developers.openai.com/api/docs/guides/conversation-state)。

## 5. 用量、错误与隐私

- 用量观测保留原始计费维度与 reported / estimated / unknown 标识。供应商 Token 指标可能存在包含关系，统一计费映射必须避免重复计数。
- 网关错误采用入站协议对应的错误 envelope；稳定的内部错误 code 包括未认证、无权限、模型不可用、能力不支持、余额不足、限流、上游不可用、结果未知等。HTTP 状态与字段在实际实现时逐协议定义。
- 限流返回可计算的 Retry-After；未知重试时间不编造。流已开始后只能按协议发送允许的错误事件或关闭连接，不能写第二份 HTTP 状态、伪造完成事件。
- 对外错误和响应头不得泄露内部账号、上游凭据、完整错误正文或其他用户的资源。默认不记录完整提示词和响应。
- 下游取消不等同上游免费；缺少最终 usage 时记录待对账，不返回或记录伪造的精确用量。

## 6. 管理 API

/api/v1/ 为控制台接口专用前缀，与模型 API 分离。采用 OpenAPI 3.1，按实现同步契约及前端类型。管理端使用独立身份会话、同源 Cookie/CSRF 策略与资源授权。金额以十进制字符串表示，时间用 UTC ISO 8601；列表使用明确分页与过滤。具体对象与字段在对应功能实现时定义，不发布虚假接口规范。

生产 SPA fallback 仅处理管理端的非 API 页面，不捕获 /v1/、/v1beta/、/api/v1/ 或健康接口。

## 7. 兼容性验收

每个已实现端点保存脱敏 fixture 和模拟上游用例，覆盖正常 JSON、SSE、多行事件、工具增量、推理、用量、上游错误、客户端取消和超时。状态型端点还须覆盖跨用户访问、绑定过期、账号故障及不可安全重试。

SDK 和 Agent 测试记录客户端名称、版本、端点、协议能力和通过范围；不以单次文本对话替代复杂工具、流式和上下文延续测试。真实付费上游只做必要的受控冒烟测试。
