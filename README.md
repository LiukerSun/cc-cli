# cc-cli

`ccc` — 给 `claude` / `codex` 用的命令行配置切换器。

把不同站点、Token、模型保存为独立 profile，终端里一条命令选择并运行——不再手动改环境变量，不再来回编辑 `~/.claude` 或 `~/.codex`。

适合这些场景：

- 同时使用官方源和第三方 relay
- 手里有多套 API Key / 多个模型
- 需要在 `claude` 和 `codex` 之间频繁切换
- 希望"选配置 → 运行"一步完成

当前主实现为 Go 版本。仓库里的 `bin/ccc` 和 `bin/ccc.ps1` 仅作兼容 wrapper，正式入口是安装后的 `ccc` 二进制。

## 为什么用它

`ccc` 把分散的几步合并为一个动作：

1. 选择 profile
2. 注入模型、认证信息和附加环境变量
3. 按需同步到 `~/.claude` 或 `~/.codex`
4. 直接启动目标 CLI

从此不用再记住：

- 该用哪个 `base_url`
- 该切哪个 `model` / `fast_model`
- 当前 shell 需要设置哪些环境变量
- `claude` / `codex` 的本地配置文件怎么改

## 三步上手

### 1. 安装

macOS / Linux:

```bash
curl -fsSL https://raw.githubusercontent.com/LiukerSun/cc-cli/main/install.sh | bash
```

Windows PowerShell:

```powershell
irm https://raw.githubusercontent.com/LiukerSun/cc-cli/main/install.ps1 | iex
```

安装路径：

- Linux / macOS: `~/.local/bin/ccc`
- Windows: `%LOCALAPPDATA%\Programs\ccc\bin\ccc.exe`

安装脚本会自动检测你的 shell 并将安装目录添加到 PATH（支持 Bash、Zsh、Fish）。如果已配置或不想自动配置，可使用：

```bash
# 跳过自动 shell 配置
curl -fsSL https://raw.githubusercontent.com/LiukerSun/cc-cli/main/install.sh | bash -s -- --no-shell-config
# 或设置环境变量
CCC_NO_SHELL_CONFIG=1 bash -c "$(curl -fsSL https://raw.githubusercontent.com/LiukerSun/cc-cli/main/install.sh)"
```

### 2. 添加配置

```bash
ccc add
```

进入交互式添加流程，内置入口：

1. `Anthropic Claude`
2. `OpenAI Codex`
3. `ZAI / ZHIPU AI`
4. `Alibaba Coding Plan`
5. `Kimi Coding Plan`
6. `Manual input`

### 3. 运行

```bash
ccc
```

行为规则：

- 无配置 → 提示先执行 `ccc add`
- 仅一个配置 → 直接运行
- 多个配置 → 打开交互选择器
- 选中后自动保存为当前 profile

## 核心功能

- **多 profile 管理** — 不同站点、模型、Token 固化为独立配置
- **一键切换并运行** — `ccc` 即可进入选择和执行流程
- **交互式选择器** — 方向键上/下或 `j`/`k`，`Enter` 运行，`q` 退出
- **快捷录入** — 支持 `ccc add openai sk-xxx [model]` 等脚本友好写法
- **Shell 补全** — 支持导出 Bash / Zsh / Fish / PowerShell 补全脚本
- **自动同步外部配置** — 运行前写入 `~/.claude` 或 `~/.codex`
- **自动检查依赖** — 缺少 `claude` / `codex` 时尝试自动安装
- **bypass 支持** — 需要时可进入最宽松运行模式

## 常见用法

```bash
ccc                              # 交互选择并运行
ccc run codex-prod               # 直接运行指定 profile
ccc run zhipu-main -- --help     # 透传参数给目标 CLI
ccc run --dry-run                # 预览执行计划
ccc run --env-only               # 仅注入环境变量
ccc run --auto-install           # 缺失目标 CLI 时允许自动安装
ccc run --auto-sync              # 兼容旧用法；默认已会同步外部配置
ccc -y                           # bypass 模式快捷入口
```

## 命令一览

```text
ccc
ccc help
ccc version
ccc -y
ccc run [profile] [--dry-run] [--env-only] [--auto-install] [--auto-sync] [-y|--bypass] [-- cli-args...]
ccc completion <bash|zsh|fish|powershell>
ccc add [<preset> <api-key> [model]] [--name ...] [--id ...]
ccc current
ccc sync [profile] [--dry-run]
ccc profile list [--json] [--show-secrets]
ccc profile add [--name ...] [--preset anthropic|openai|zhipu|alibaba|kimi] --api-key ...
ccc profile update <profile> [--preset ...] [--model ...]
ccc profile use <profile>
ccc profile delete <profile>
ccc paths [--json]
ccc config path
ccc config show [--show-secrets]
ccc config migrate
ccc doctor
ccc upgrade [--version <semver>] [--dry-run]
```

> - `ccc` 等价于"直接开始运行流程"
> - `ccc -y` 是 `ccc run -y` 的顶层快捷方式
> - `ccc run foo -- --help` 会把 `--help` 透传给目标 CLI
> - `ccc --help` / `ccc --version` 作为兼容别名保留

## Shell 补全

生成补全脚本：

```bash
ccc completion bash
ccc completion zsh
ccc completion fish
ccc completion powershell
```

常见用法：

```bash
# Bash
source <(ccc completion bash)

# Zsh
source <(ccc completion zsh)

# Fish
ccc completion fish | source
```

PowerShell:

```powershell
ccc completion powershell | Out-String | Invoke-Expression
```

## 添加配置

### 交互模式

```bash
ccc add
```

内置入口：

1. **Anthropic Claude** — 输入 API Key，选择主模型和快速模型
2. **OpenAI Codex** — 输入 Base URL、API Key、模型
3. **ZAI / ZHIPU AI** — 输入 API Key，自动拉取模型列表，选择主模型和快速模型
4. **Alibaba Coding Plan** — 输入 API Key，自动拉取模型列表，选择主模型和快速模型
5. **Kimi Coding Plan** — 输入 API Key，默认使用 `https://api.kimi.com/coding/`，且模型列表仅提供 `K2.6-code-preview`
6. **Manual input** — 手动指定 `claude` 或 `codex`、Base URL、模型等字段

在线拉取模型失败时会自动回退到内置模型列表，不会因接口请求卡死。

当前内置的 Alibaba Coding Plan 回退模型列表：

- 千问：`qwen3.6-plus`、`qwen3.5-plus`、`qwen3-max-2026-01-23`、`qwen3-coder-next`、`qwen3-coder-plus`
- 智谱：`glm-5`、`glm-4.7`
- Kimi：`kimi-k2.5`
- MiniMax：`MiniMax-M2.5`
交互录入 API Key 时默认不回显。

### 快捷模式

适合脚本、临时录入或批量配置：

```bash
ccc add openai sk-xxx
ccc add openai sk-xxx gpt-5.4
ccc add zhipu sk-xxx glm-5
ccc add alibaba sk-xxx qwen3.6-plus
ccc add kimi sk-xxx K2.6-code-preview
ccc add anthropic sk-ant-xxx claude-3-7-sonnet
```

支持这些 preset：

| Preset | 别名 |
|--------|------|
| `anthropic` | `claude` |
| `openai` | `codex`、`gpt` |
| `zhipu` | `zai`、`glm` |
| `alibaba` | `qwen`、`dashscope`、`tongyi` |
| `kimi` | `moonshot` |

Kimi preset 走 Claude 兼容入口 `https://api.kimi.com/coding/`。
如果遇到 Kimi 文档里提到的 `tool_search` 相关 400，可对该 profile 追加：

```bash
ccc profile update my-kimi --env ENABLE_TOOL_SEARCH=false
```

### 精细控制

完全自定义时，直接使用 profile 命令：

```bash
ccc profile add \
  --name "My Relay" \
  --command codex \
  --provider custom \
  --base-url https://relay.example.com/v1 \
  --api-key sk-xxx \
  --model gpt-5.4
```

日常维护：

```bash
ccc profile list
ccc current
ccc profile use my-relay
ccc profile update my-relay --model gpt-5.4-mini
ccc profile update my-relay --env OPENAI_ORG=demo
ccc profile duplicate my-relay --name "My Relay Backup"
ccc profile export my-relay --output relay.json
ccc profile import --input relay.json
ccc profile delete my-relay
```

`ccc profile list --json` 和 `ccc config show` 默认会对 `api_key` 脱敏；只有显式加 `--show-secrets` 才会输出原值。

## 运行与选择

交互选择器：

- 方向键上/下或 `j`/`k` 移动
- `Enter` 运行
- `q` 退出
- 终端不支持方向键原始模式时，自动回退到数字选择

```bash
ccc                          # 交互选择
ccc run codex-prod           # 直接运行
ccc run zhipu-main -- --help # 透传参数
ccc run --dry-run            # 预览
ccc run --env-only           # 仅注入环境变量
ccc run --auto-install       # 缺失 codex/claude CLI 时自动安装
ccc run --auto-sync          # 兼容旧用法；默认已同步 ~/.codex 或 ~/.claude
```

`--dry-run` 展示的执行计划包括：

- 选中的 profile
- 实际运行的命令
- 缺失 CLI 时是直接失败还是自动安装
- 是否会同步外部配置
- 目标 CLI 收到的参数
- 实际注入的环境变量

`ccc run` 默认不会自动安装目标 CLI，但会在 profile 启用同步时先写入外部配置。
需要显式控制时：

- 用 `ccc run --auto-install ...` 允许自动安装缺失的 `claude` / `codex`
- `ccc run --auto-sync ...` 作为兼容参数仍可使用，但已不是必需
- 或直接使用 `ccc sync` 单独执行配置同步

## `-y` / bypass

直接进入 bypass 模式：

```bash
ccc -y
ccc run -y
ccc run my-profile -y
ccc upgrade --check
```

不同 CLI 的处理方式：

- **claude** — 注入 `CLAUDE_SKIP_PERMISSIONS=1`、`IS_SANDBOX=1`，并追加 `--dangerously-skip-permissions`
- **codex** — 追加 `--yolo`

> 此功能是为兼容当前 CLI 的实际行为，请确认了解风险后再使用。

## 外部配置同步

默认情况下，profile 运行前会自动同步外部配置。

**claude** profile 写入 `~/.claude/settings.json`，同步字段：

- `ANTHROPIC_MODEL`
- `ANTHROPIC_SMALL_FAST_MODEL`
- `CLAUDE_CODE_MODEL`
- `CLAUDE_CODE_SMALL_MODEL`
- `CLAUDE_CODE_SUBAGENT_MODEL`
- `permissions.deny += <profile.sync_deny_permissions>`

如需显式追加 deny 规则，可在新增或更新 profile 时传入：

```bash
ccc profile add ... --deny-permission 'Agent(Explore)'
ccc profile update my-claude --deny-permission 'Bash(rm -rf)'
```

**codex** profile 写入 `~/.codex/config.toml` 和 `~/.codex/auth.json`，同步字段：

- `model_provider = "codex"`
- `model = "<当前模型>"`
- `[model_providers.codex].base_url`
- `[model_providers.codex].wire_api = "responses"`
- `OPENAI_API_KEY`

新增或更新 profile 时可加 `--no-sync` 关闭同步。

单独执行同步：

```bash
ccc sync
ccc sync --dry-run
ccc sync my-profile
```

兼容旧入口：

```bash
ccc --list
ccc --add openai sk-xxx gpt-5.4
ccc --delete my-profile
ccc --current
ccc -e my-profile
```

## 配置与目录

Go 版本使用标准目录布局。

**Linux / macOS:**

| 用途 | 路径 |
|------|------|
| Binary | `~/.local/bin/ccc` |
| Config | `~/.config/ccc/config.json` |
| Data | `~/.local/share/ccc` |
| Cache | `~/.cache/ccc` |
| State | `~/.local/state/ccc` |

**Windows:**

| 用途 | 路径 |
|------|------|
| Binary | `%LOCALAPPDATA%\Programs\ccc\bin\ccc.exe` |
| Config | `%APPDATA%\ccc\config.json` |
| Data | `%LOCALAPPDATA%\ccc\data` |
| Cache | `%LOCALAPPDATA%\ccc\cache` |
| State | `%LOCALAPPDATA%\ccc\state` |

兼容读取旧配置：`~/.ccc/config.json`、`~/.cc-config.json`

迁移到新路径：

```bash
ccc config migrate
```

排查问题：

```bash
ccc paths
ccc config path
ccc config show
ccc doctor
```

## 运行时依赖

`ccc` 本体是 Go 二进制，目标 CLI 仍依赖 Node.js / npm。

自动安装时的要求：

- **claude**: Node.js `>= 18.0.0`
- **codex**: Node.js `>= 16.0.0`

目标命令不存在时，`ccc run` 默认会直接报错；只有显式加 `--auto-install` 才会尝试自动安装。

## 升级

```bash
ccc upgrade --dry-run             # 预览
ccc upgrade                       # 执行升级
ccc upgrade --version 2.5.6       # 升级到指定版本
```

- Linux / macOS / Windows 均支持原地升级
- Windows 升级后如果当前终端仍显示旧版本，重开终端即可

## 开发

```bash
make build
make test
```

等价于：

```bash
go test ./...
bash tests/test.sh
```

仓库中的 `bin/ccc` 和 `bin/ccc.ps1` 仅用于兼容旧入口，不再是主实现。

**版本发布**以 `VERSION` 文件为准，流程：

1. 更新 `VERSION`
2. 提交并合并到 `main`
3. 创建并推送同版本 tag（如 `v2.2.1`）
4. GitHub Actions release workflow 校验 tag 与 `VERSION` 一致后执行 GoReleaser

本地校验：

```bash
bash tests/check_version.sh
bash tests/check_version.sh v2.2.1
```
