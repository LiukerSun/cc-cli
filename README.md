# cc-cli

`ccc` 是一个给 `claude` / `codex` 用的命令行配置切换器。

你可以把不同站点、不同 Token、不同模型保存成多个 profile，然后在终端里直接选择并运行，不用再手改环境变量，也不用来回改 `~/.claude` 或 `~/.codex`。

如果你经常遇到下面这些情况，它就是为你写的：

- 同时用官方源和第三方 relay
- 一个人手里有多套 API Key / 多个模型
- 需要在 `claude` 和 `codex` 之间快速切换
- 想把“切配置再运行”变成一次命令完成

当前主实现已经是 Go 版本。仓库里的 `bin/ccc` 和 `bin/ccc.ps1` 只是兼容 wrapper，正式入口是安装后的 `ccc` 二进制。

## 为什么更省事

`ccc` 把原本分散的几步合并成一个动作：

- 选择 profile
- 注入当前 profile 的模型、认证和附加环境变量
- 按需同步到 `~/.claude` 或 `~/.codex`
- 直接启动目标 CLI

也就是说，你不用再自己记住：

- 这次该切哪个 `base_url`
- 这次该用哪个 `model` / `fast_model`
- 当前 shell 里该设置哪些环境变量
- `claude` 和 `codex` 的本地配置文件该怎么改

## 三步开始

### 1. 安装

macOS / Linux:

```bash
curl -fsSL https://raw.githubusercontent.com/LiukerSun/cc-cli/main/install.sh | bash
```

Windows PowerShell:

```powershell
irm https://raw.githubusercontent.com/LiukerSun/cc-cli/main/install.ps1 | iex
```

默认安装路径：

- Linux / macOS: `~/.local/bin/ccc`
- Windows: `%LOCALAPPDATA%\Programs\ccc\bin\ccc.exe`

### 2. 添加一个配置

```bash
ccc add
```

你会进入交互式添加流程，当前内置入口有：

1. `ZAI / ZHIPU AI`
2. `Alibaba Coding Plan`
3. `OpenAI Codex`
4. `Manual input`

### 3. 直接运行

```bash
ccc
```

默认行为很简单：

- 没有配置时，提示先执行 `ccc add`
- 只有一个配置时，直接运行该 profile
- 有多个配置时，在交互终端里打开选择器
- 选中的 profile 会自动保存成当前 profile

## 核心功能

- 多 profile 管理：把不同站点、模型、Token 固化成独立配置
- 一键切换并运行：`ccc` 直接进入选择和执行流程
- 交互式选择器：支持方向键上 / 下，也支持 `j` / `k`
- 快捷录入：保留 `ccc add openai sk-xxx [model]` 这类脚本友好写法
- 自动同步外部配置：运行前写入 `~/.claude` 或 `~/.codex`
- 自动检查依赖：缺少 `claude` / `codex` 时尝试自动安装
- bypass 支持：需要时直接进入最宽松运行模式

## 常见使用方式

```bash
ccc
ccc run codex-prod
ccc run zhipu-main -- --help
ccc run --dry-run
ccc run --env-only
ccc -y
```

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
