# peacesheep的个人主页-后端



## 使用wire生成代码

```bash
go install github.com/google/wire/cmd/wire@latest
cd biz/container
wire
```

## 运行

```bash
go run .
```

服务启动时会读取 `conf/config.yaml`。JWT 密钥必须至少 32 字节，且不能使用示例值
`secret`；缺失或弱密钥会直接终止启动。密钥读取优先级如下：

1. `PERSONAL_PAGE_JWT_KEY`
2. `http.jwt_key`
3. 兼容旧配置的 `http.secret_key`

Session 签名密钥可以通过 `PERSONAL_PAGE_SESSION_KEY` 或
`http.session_key` 单独配置；未配置时会从已经校验过的 JWT 密钥做域分离派生，
不会使用硬编码默认值。AI 访客伪匿名密钥可通过 `AICHAT_IDENTITY_KEY` 单独配置，
未配置时同样从 JWT 密钥做另一用途的域分离派生。

## AI 对话安全配置

`POST /api/aichat` 接受前端的 `X-AI-Visitor-ID`。服务只保存它的 HMAC 摘要，
不会把原始访客 ID 或 IP 发送给 Dify 或写入用量表。登录用户与匿名访客使用相互
隔离的 Dify `user`。反向代理应覆盖并传入可信的 `X-Real-IP`；服务不会使用可由
客户端任意拼接的 `X-Forwarded-For`。只有 `AICHAT_TRUST_PROXY_HEADERS=true` 时才会
读取 `X-Real-IP`；默认只使用连接远端地址，避免可绕过反向代理的入口伪造 IP。

| 环境变量 | 默认值 | 说明 |
| --- | ---: | --- |
| `HTTP_MAX_REQUEST_BODY_BYTES` | `33554432` | Hertz 全局请求体上限（32 MiB；旧的大文件接口如有需要可显式调高） |
| `AICHAT_MAX_REQUEST_BODY_BYTES` | `65536` | AI 对话 JSON 请求体上限（64 KiB） |
| `AICHAT_MAX_MESSAGE_RUNES` | `4000` | 单条消息 Unicode 字符上限 |
| `AICHAT_REQUESTS_PER_MINUTE` | `6` | 每个伪匿名身份且每个 IP、每个后端进程的分钟请求上限 |
| `AICHAT_MAX_CONCURRENT` | `1` | 每个伪匿名身份且每个 IP、每个后端进程的并发上限 |
| `AICHAT_DAILY_REQUEST_BUDGET` | `50` | 每个伪匿名身份且每个 IP 的每日请求上限（Asia/Shanghai） |
| `AICHAT_TRUST_PROXY_HEADERS` | `false` | 是否信任由受控反向代理覆写的 `X-Real-IP` |

以上值也可写在 `aichat.max-request-body-bytes`、
`aichat.max-message-runes`、`aichat.requests-per-minute`、
`aichat.max-concurrent`、`aichat.daily-request-budget` 和
`aichat.trust-proxy-headers`。环境变量优先，所有限额必须为正整数，否则服务拒绝
启动。日预算通过 `ai_usage_daily_quota_entities` 表按日期、伪匿名身份和 IP 在同一
事务中原子预占，因此多副本并发不会穿透上限；上游失败或客户端断开也保留本次
计数。数据库不可用时 AI 请求会 fail closed，不会继续消耗上游额度。

AI 对话错误使用标准 HTTP 状态：参数错误为 `400`，请求体过大为 `413`，分钟、
并发或日预算达到上限为 `429`（并返回 `Retry-After`），配置或预算数据库暂时不可用
为 `503`。正常流式响应仍为 `200 text/event-stream`。

## 功能

1. 登陆
2. 上传下载文件
3. 简易聊天
