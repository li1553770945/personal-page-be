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

## 生产部署

生产发布不再使用 GitHub Actions、Docker Hub、`latest` 或 Keel。代码已经提交后，
Windows PowerShell 在仓库根目录执行一条命令即可完成源码推送和生产发布：

```powershell
.\scripts\deploy-backend.ps1 -PushSource
```

脚本会依次完成：

1. 要求工作树干净且位于 `master`，拒绝远端分支领先或分叉。
2. 运行 `go test ./...` 和 `go vet ./...`；只有显式传入 `-PushSource` 时才把本地已提交
   内容推送到 `origin/master`，否则要求两端已经一致。
3. 构建 `linux/amd64` 镜像并推送完整 Git SHA 标签到
   `ccr.ccs.tencentyun.com/littlehorse/personal-page-be`；已存在的 SHA 标签绝不覆盖。
   Go builder 使用与官方镜像相同摘要的国内镜像并固定 `sha256`；若用
   `DOCKER_GO_BUILDER_IMAGE` 覆盖，也必须包含明确的 `@sha256:...`。
4. 验证 Kubernetes 中已有的 `ccr-tencent` pull secret；首次显式传入
   `-BootstrapPullSecret` 时，才从本机 Docker credential store 读取 CCR 登录并经
   SSH 标准输入安全创建 Secret，不把凭据写进仓库或命令行参数。
5. 通过 SSH 将 Deployment 显式更新到 `SHA tag@digest`，等待 rollout，验证两个 Pod、
   NodePort 和公网 `/api/ping`；失败时仅在没有并发发布覆盖的前提下恢复上一完整镜像。

首次使用前先完成 CCR 登录并确认 SSH 主机指纹：

```powershell
docker login ccr.ccs.tencentyun.com
ssh ubuntu@124.223.181.152
```

随后用一次显式 bootstrap 完成首次发布：

```powershell
.\scripts\deploy-backend.ps1 -PushSource -BootstrapPullSecret
```

`-BootstrapPullSecret` 会把当前本机 CCR 写凭据存入 Kubernetes Secret，因此只用于
首次初始化或凭据轮换；如果 CCR 提供独立只读拉取凭据，应优先用它初始化 Secret。
后续日常发布不再读取或修改该 Secret。

如果镜像已经推送、但部署阶段中断，可继续发布当前提交：

```powershell
.\scripts\deploy-backend.ps1 -Resume
```

回退到仍保留在 CCR 中的旧提交：

```powershell
.\scripts\deploy-backend.ps1 -RollbackTag <完整的40位Git-SHA>
```

只查看将要发布的标签和目标，不执行任何外部修改：

```powershell
.\scripts\deploy-backend.ps1 -DryRun
```

发布前必须先提交代码；脚本不会自动提交未审查的工作树。镜像回退也不会回退数据库
结构，因此生产迁移必须保持向后兼容。腾讯云 CCR 个人版单仓库版本数有限，应定期清理
不再用于当前版本及回退窗口的旧 SHA 标签，但不得覆盖或复用已有 SHA 标签。

## 功能

1. 登陆
2. 上传下载文件
3. 简易聊天
