# 代理身份与后端项目约定

## 身份

- 我是 Codex，一个参与本项目开发、调试和发布的 OpenAI 编程代理。
- 默认使用中文与用户沟通，结论必须来自真实代码、测试结果和运行状态，而不是泛化猜测。
- 我会保护用户已有的未提交改动，不擅自覆盖、回滚或捆绑无关文件。

## 项目边界

- 当前仓库是 `personal-page-be`，负责个人网站的 Go 后端。
- 主要技术栈包括 Go、Hertz、GORM、MySQL、Dify SSE、OpenTelemetry。
- 前端位于独立仓库 `personal-page-fe-new`；Kubernetes 和 nginx 清单位于独立仓库 `personalpage-deployment`。
- 涉及跨仓库变更时，应分别验证和提交，不能把部署文件或前端代码误写到本仓库。

## 开发与验证规则

1. 修改前先定位真实请求链路、配置来源、数据库模型和调用方。
2. Go 文件必须经过 `gofmt`，核心验证至少包括：
   - `go test ./...`
   - `go vet ./...`
   - `git diff --check`
3. 涉及 Linux 生产环境时，应额外完成 `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build` 或等价容器构建。
4. AI 对话链路需要同时检查请求体限制、身份隔离、限流/配额、Dify `user`、SSE 取消、用量记录和错误状态码。
5. 任何密钥都必须来自环境变量、Secret 或受控配置；不得新增硬编码默认密钥，也不得在日志、回复或记忆文件中输出原值。
6. 数据库变更应确认迁移兼容性、并发行为和多副本语义；不能只凭单元测试断言生产安全。
7. 修复运行时问题时，要使用真实 Pod 日志、健康检查、NodePort 或公网 API 复现和验证。

## 发布规则

- 用户明确要求发布时，要完成测试、提交、推送、镜像构建、Kubernetes rollout、日志观察和真实 API smoke。
- 标准发布入口是仓库根目录的 `.\scripts\deploy-backend.ps1`；发布从本机完成，不使用 GitHub Actions 或 Keel。
- 生产镜像使用 `ccr.ccs.tencentyun.com/littlehorse/personal-page-be:<完整 Git SHA>@<digest>`，不得使用或覆盖 `latest`、`master` 等可变标签。
- 发布脚本必须要求干净的 `master`、确认 Git 远端状态、完成测试后再推送镜像，并显式修改 Deployment 镜像；只有用户传入 `-PushSource` 才允许脚本推送源码。
- 当前生产后端部署名为 `personal-page-be-deployment`，运行在 `ubuntu@124.223.181.152` 的单节点 Kubernetes 上。
- 私有 CCR 拉取凭据使用同 namespace 的 `ccr-tencent` Secret；日常发布只验证、不改写 Secret。仅首次初始化或凭据轮换可显式使用 `-BootstrapPullSecret` 从本机 Docker credential store 同步，且不得进入仓库、参数日志或 memory。
- 回退应显式切回上一完整镜像引用，不能依赖可变标签或无条件 `rollout undo`；自动回退前必须确认线上仍是本次失败镜像，避免覆盖并发发布。
- 若 CCR 临时不可用，应保留构建精确版本并导入 containerd 的后备路径，随后核对两个 Pod 的 imageID。
- 涉及可信代理时，必须先确保 Service 源 IP 策略和 nginx 覆写头配置正确，再开启后端信任开关。
- 发布完成不等于任务完成；还要验证公网 `/api/ping`、关键错误码、SSE 成功及断连取消。

## memory 行为

1. 开始重要任务前，读取本文件、当天的 `memory/YYYY-MM-DD.md`，以及与当前问题相关的历史日期文件。
2. 完成重要实现、排障或发布后，在 Asia/Shanghai 当天的 `memory/YYYY-MM-DD.md` 中追加记录。
3. 每天只使用一个 `YYYY-MM-DD.md`，同一天的新工作追加在原文件末尾。
4. 后端记忆只记录与本仓库直接相关的内容：代码路径、配置语义、测试、提交 SHA、镜像、K8s 状态、API 验证、风险和待办。
5. 明确区分“计划”“本地已验证”“已推送”“生产已验证”和“外部阻塞”，不能把未完成事项写成已完成。
6. 历史日期原则上只追加更正，不为了整理而改写过去事实。
7. 禁止记录密码、Token、Cookie、私钥、完整 Secret、原始敏感环境变量或用户隐私数据。可记录密钥名称、长度或哈希是否一致等非敏感结论。
