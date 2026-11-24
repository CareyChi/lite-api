# API 契约与协议兼容要求

文档描述计划接口与保真约束；除标记为 S0 的接口外，路径尚未实现。

## 1. 当前接口（S0 已实现）

| 方法 | 路径 | 目的 |
| --- | --- | --- |
| GET | /healthz | 进程存活探针，返回 JSON status=ok |
| GET | /api/v1/system/info | 工程状态，返回 name、status=scaffold、api_version=v1 |

S0 尚无用户登录、鉴权、计费与 AI 上游请求；不得将上述健康接口当成生产就绪探针。

## 2. 计划的原生网关入口

| 协议 | 方法与路径 | 最早阶段 |
| --- | --- | --- |
| OpenAI | GET /v1/models | S1 |
| OpenAI | POST /v1/chat/completions | S1 |
| OpenAI | POST /v1/responses | S4 |
| Anthropic | POST /v1/messages | S2 |
| Gemini | POST /v1beta/models/{model}:generateContent | S2 |
| Gemini | POST /v1beta/models/{model}:streamGenerateContent | S2 |

流式 Gemini 入口需兼容相应查询参数（例如 alt=sse）；各协议版本及适用模型按能力矩阵单独管理。后续管理 API 统一采用 /api/v1/ 前缀，与上游兼容入口分离。

## 3. 鉴权与错误

- 入站 OpenAI 风格使用 Bearer API Key；Anthropic 与 Gemini 可支持其原生认证头/查询方式，但必须归一到网关颁发的客户端凭据，不把上游私钥暴露给客户端。
- 对外错误优先符合请求入口的原生错误格式；网关自定义错误有稳定 code、message 和 request_id，不能在 SSE 中混用不兼容的 JSON 错误帧。
- 未授权、余额不足、速率限制、上游不可用必须可明确区分。
- 禁止把上游认证头、Cookie、后端密钥或内部账号 ID 透传给下游。

## 4. 兼容性与保真

- 原生透传优先，保留合法的厂商扩展字段和响应事件。
- 工具调用参数与增量事件不得丢失；思考块、签名、缓存控制语义不等价时明确返回不支持。
- Token 使用量的来源标记必须保留（上游报告 / 网关估算）；不伪造缓存命中。
- SSE 必须保持事件顺序、及时 flush，客户端断开时取消上游请求并释放租约。
- 只有在响应尚未开始且请求满足重试规则时，才能换上游重试。
- 对每种入口定义基于真实协议样例的非流式、流式、取消、工具调用和错误用例。

## 5. 后续接口设计规则

- 使用 OpenAPI 3.1 管理控制台接口契约并生成前端类型；模型兼容接口以实际供应商协议测试为准。
- 版本化接口与数据迁移同步审查；不得私自将不兼容的结构变化混入已有路径。
- 详细 HTTP 请求/响应模型应在实现对应阶段时补充，不提前提供与真实行为不一致的伪规范。
