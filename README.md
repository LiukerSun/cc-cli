# cc-cli

`ccc` 是一个面向终端用户的模型切换器。

它解决的是一个很具体的问题：你已经装了 `claude` 或 `codex`，也有多套官方源、转发地址、不同模型和不同 API Key，但每次切换都要改环境变量、改本地配置、再重新执行命令，过程很碎。`ccc` 把这件事压缩成了“选一个 profile，然后直接开跑”。

适合的场景：

- 你在 `claude` 和 `codex` 之间来回切
- 你同时用官方源和第三方 relay
- 你希望把 `base_url`、`api_key`、`model`、`fast_model` 固化成多个 profile
- 你不想每次手改 `~/.claude` 或 `~/.codex`

当前主实现已经是 Go 版本。仓库里的 `bin/ccc` 和 `bin/ccc.ps1` 只是兼容 wrapper，正式入口是安装后的 `ccc` 二进制。

## 30 秒上手

安装：

macOS / Linux:

```bash
curl -fsSL https://raw.githubusercontent.com/LiukerSun/cc-cli/main/install.sh | bash
```

Windows PowerShell:

```powershell
irm https://raw.githubusercontent.com/LiukerSun/cc-cli/main/install.ps1 | iex
```

添加一个配置：

```bash
ccc add
```

直接运行：

```bash
ccc
```

默认行为：

- 没有配置时，提示先执行 `ccc add`
- 只有一个配置时，直接运行该 profile
- 有多个配置时，在交互终端里打开选择器
- 选中的 profile 会自动保存成当前 profile

## 这个工具会帮你做什么

`ccc` 运行一个 profile 时，通常会顺手做完这些事情：

- 选择要执行的目标命令：`claude` 或 `codex`
- 注入当前 profile 对应的模型和认证信息
- 在需要时同步外部配置到 `~/.claude` 或 `~/.codex`
- 目标 CLI 缺失时，先检查依赖并尝试自动安装
- 需要时进入 bypass 模式

你不需要再手动记住：

- 这次应该切哪个 `base_url`
- 这次要用哪个模型
- 当前终端里该设置哪些环境变量
- `codex` 和 `claude` 的本地配置文件各自该怎么改

## 核心体验

- `ccc` 直接进入运行流程
- 默认支持交互式 profile 选择
- 支持方向键上 / 下，也支持 `j` / `k`
- `ccc add` 无参数进入交互式添加
- 保留 `ccc add openai sk-xxx [model]` 这类快捷写法
- 默认自动同步 `~/.claude` 或 `~/.codex`
- 缺少 `claude` / `codex` 时会尝试自动安装

## 常用命令

```text
ccc
ccc help
ccc version
ccc -y
ccc run [profile-id-or-name] [--dry-run] [--env-only] [-y|--bypass] [-- cli-args...]
ccc add [<preset> <api-key> [model]] [--name ...] [--id ...]
ccc current
ccc sync [profile-id-or-name] [--dry-run]
ccc profile list [--json]
ccc profile add [--name ...] [--preset anthropic|openai|zhipu|alibaba] --api-key ...
ccc profile update <profile-id-or-name> [--preset anthropic|openai|zhipu|alibaba] [--model ...]
ccc profile use <profile-id-or-name>
ccc profile delete <profile-id-or-name>
ccc paths [--json]
ccc config path
ccc config show
ccc config migrate
ccc doctor
ccc upgrade [--version <semver>] [--dry-run]
```

补充说明：

- `ccc` 等价于“直接开始运行流程”
- `ccc -y` 是 `ccc run -y` 的顶层快捷方式
- `ccc run foo -- --help` 会把 `--help` 透传给目标 CLI
- `ccc --help` 和 `ccc --version` 作为兼容别名保留

## 添加配置

### 交互模式

```bash
ccc add
```

当前内置入口：

1. `ZAI / ZHIPU AI`
2. `Alibaba Coding Plan`
3. `OpenAI Codex`
4. `Manual input`

各入口行为：

- `ZAI / ZHIPU AI`：输入 API Key，自动拉取模型列表，选择主模型和快速模型
- `Alibaba Coding Plan`：输入 API Key，自动拉取模型列表，选择主模型和快速模型
- `OpenAI Codex`：输入 Base URL、API Key、模型
- `Manual input`：手动指定 `claude` 或 `codex`、Base URL、模型等字段

如果在线拉取模型失败，会自动回退到内置模型列表，不会卡死在接口请求上。

### 快捷模式

适合脚本、临时录入或一次性批量配置：

```bash
ccc add openai sk-xxx
ccc add openai sk-xxx gpt-5.4
ccc add zhipu sk-xxx glm-5
ccc add alibaba sk-xxx qwen3.5-plus
ccc add anthropic sk-ant-xxx claude-3-7-sonnet
```

常见 preset：

- `anthropic`
- `openai`
- `zhipu`
- `alibaba`

也支持这些别名：

- `claude` -> `anthropic`
- `codex` / `gpt` -> `openai`
- `zai` / `glm` -> `zhipu`
- `qwen` / `dashscope` / `tongyi` -> `alibaba`

### 精细控制

如果你想完全自己定义，也可以直接走 profile 命令：

```bash
ccc profile add \
  --name "My Relay" \
  --command codex \
  --provider custom \
  --base-url https://relay.example.com/v1 \
  --api-key sk-xxx \
  --model gpt-5.4
```

常用维护命令：

```bash
ccc profile list
ccc current
ccc profile use my-relay
ccc profile update my-relay --model gpt-5.4-mini
ccc profile update my-relay --env OPENAI_ORG=demo
ccc profile delete my-relay
```

## 运行与选择

默认选择器支持：

- 方向键上 / 下
- `j` / `k`
- `Enter` 运行
- `q` 退出

如果当前终端不支持方向键原始模式，会自动回退到数字选择。

常见例子：

```bash
ccc
ccc run codex-prod
ccc run zhipu-main -- --help
ccc run --dry-run
ccc run --env-only
```

`--dry-run` 会展示最终执行计划，包括：

- 选中的 profile
- 实际运行的命令
- 是否会同步外部配置
- 目标 CLI 会收到哪些参数
- 实际注入的环境变量

## `-y` / bypass

如果你希望直接进入 bypass 模式：

```bash
ccc -y
ccc run -y
ccc run my-profile -y
```

不同命令的处理方式不同：

- `claude`：注入 `CLAUDE_SKIP_PERMISSIONS=1`、`IS_SANDBOX=1`，并追加 `--dangerously-skip-permissions`
- `codex`：追加 `--dangerously-bypass-approvals-and-sandbox`

这部分逻辑是为了兼容当前 CLI 的实际行为，不建议在不了解风险的环境里滥用。

## 外部配置同步

默认情况下，profile 会在运行前自动同步外部配置。

`claude` profile 会写入：

- `~/.claude/settings.json`

同步的关键字段包括：

- `ANTHROPIC_MODEL`
- `ANTHROPIC_SMALL_FAST_MODEL`
- `CLAUDE_CODE_MODEL`
- `CLAUDE_CODE_SMALL_MODEL`
- `CLAUDE_CODE_SUBAGENT_MODEL`
- `permissions.deny += Agent(Explore)`

`codex` profile 会写入：

- `~/.codex/config.toml`
- `~/.codex/auth.json`

同步的关键字段包括：

- `model_provider = "codex"`
- `model = "<当前模型>"`
- `[model_providers.codex].base_url`
- `[model_providers.codex].wire_api = "responses"`
- `OPENAI_API_KEY`

如果你不想同步外部配置，可以在新增或更新 profile 时使用 `--no-sync`。

也可以单独执行：

```bash
ccc sync
ccc sync --dry-run
ccc sync my-profile
```

## 配置与目录

当前 Go 版本使用标准目录布局。

Linux / macOS:

- Binary: `~/.local/bin/ccc`
- Config: `~/.config/ccc/config.json`
- Data: `~/.local/share/ccc`
- Cache: `~/.cache/ccc`
- State: `~/.local/state/ccc`

Windows:

- Binary: `%LOCALAPPDATA%\Programs\ccc\bin\ccc.exe`
- Config: `%APPDATA%\ccc\config.json`
- Data: `%LOCALAPPDATA%\ccc\data`
- Cache: `%LOCALAPPDATA%\ccc\cache`
- State: `%LOCALAPPDATA%\ccc\state`

兼容读取旧配置：

- `~/.ccc/config.json`
- `~/.cc-config.json`

如需迁移到新路径：

```bash
ccc config migrate
```

排查时可用：

```bash
ccc paths
ccc config path
ccc config show
ccc doctor
```

## 运行时依赖

`ccc` 本体是 Go 二进制，但目标 CLI 仍然依赖 Node.js / npm。

自动安装目标 CLI 时，当前要求：

- `claude`：Node.js `>= 18.0.0`
- `codex`：Node.js `>= 16.0.0`

如果目标命令不存在，`ccc` 会在运行前检查依赖并尝试自动安装。

## 升级

```bash
ccc upgrade --dry-run
ccc upgrade
ccc upgrade --version 2.5.6 --dry-run
```

说明：

- Unix 平台支持原地升级当前二进制
- Windows 当前建议重新执行 `install.ps1`

## 开发

本地开发至少执行：

```bash
make build
make test
```

等价命令：

```bash
go test ./...
bash tests/test.sh
```

仓库中的 `bin/ccc` 和 `bin/ccc.ps1` 仅用于兼容旧入口，不再是主实现。

版本发布以 `VERSION` 文件为准。

基本流程：

1. 更新 `VERSION`
2. 提交改动并合并到 `main`
3. 创建并推送同版本 tag，例如 `v2.2.1`
4. GitHub Actions 中的 release workflow 会校验 tag 与 `VERSION` 一致，然后执行 GoReleaser

本地可先检查：

```bash
bash tests/check_version.sh
bash tests/check_version.sh v2.2.1
```
