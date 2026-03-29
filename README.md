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

### 2. 添加配置

```bash
ccc add
```

进入交互式添加流程，内置入口：

1. `ZAI / ZHIPU AI`
2. `Alibaba Coding Plan`
3. `OpenAI Codex`
4. `Manual input`

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
ccc -y                           # bypass 模式快捷入口
```

## 命令一览

```text
ccc
ccc help
ccc version
ccc -y
ccc run [profile] [--dry-run] [--env-only] [-y|--bypass] [-- cli-args...]
ccc add [<preset> <api-key> [model]] [--name ...] [--id ...]
ccc current
ccc sync [profile] [--dry-run]
ccc profile list [--json]
ccc profile add [--name ...] [--preset anthropic|openai|zhipu|alibaba] --api-key ...
ccc profile update <profile> [--preset ...] [--model ...]
ccc profile use <profile>
ccc profile delete <profile>
ccc paths [--json]
ccc config path
ccc config show
ccc config migrate
ccc doctor
ccc upgrade [--version <semver>] [--dry-run]
```

> - `ccc` 等价于"直接开始运行流程"
> - `ccc -y` 是 `ccc run -y` 的顶层快捷方式
> - `ccc run foo -- --help` 会把 `--help` 透传给目标 CLI
> - `ccc --help` / `ccc --version` 作为兼容别名保留

## 添加配置

### 交互模式

```bash
ccc add
```

内置入口：

1. **ZAI / ZHIPU AI** — 输入 API Key，自动拉取模型列表，选择主模型和快速模型
2. **Alibaba Coding Plan** — 输入 API Key，自动拉取模型列表，选择主模型和快速模型
3. **OpenAI Codex** — 输入 Base URL、API Key、模型
4. **Manual input** — 手动指定 `claude` 或 `codex`、Base URL、模型等字段

在线拉取模型失败时会自动回退到内置模型列表，不会因接口请求卡死。

### 快捷模式

适合脚本、临时录入或批量配置：

```bash
ccc add openai sk-xxx
ccc add openai sk-xxx gpt-5.4
ccc add zhipu sk-xxx glm-5
ccc add alibaba sk-xxx qwen3.5-plus
ccc add anthropic sk-ant-xxx claude-3-7-sonnet
```

支持这些 preset：

| Preset | 别名 |
|--------|------|
| `anthropic` | `claude` |
| `openai` | `codex`、`gpt` |
| `zhipu` | `zai`、`glm` |
| `alibaba` | `qwen`、`dashscope`、`tongyi` |

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
ccc profile delete my-relay
```

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
```

`--dry-run` 展示的执行计划包括：

- 选中的 profile
- 实际运行的命令
- 是否会同步外部配置
- 目标 CLI 收到的参数
- 实际注入的环境变量

## `-y` / bypass

直接进入 bypass 模式：

```bash
ccc -y
ccc run -y
ccc run my-profile -y
```

不同 CLI 的处理方式：

- **claude** — 注入 `CLAUDE_SKIP_PERMISSIONS=1`、`IS_SANDBOX=1`，并追加 `--dangerously-skip-permissions`
- **codex** — 追加 `--dangerously-bypass-approvals-and-sandbox`

> 此功能是为兼容当前 CLI 的实际行为，请确认了解风险后再使用。

## 外部配置同步

默认情况下，profile 运行前会自动同步外部配置。

**claude** profile 写入 `~/.claude/settings.json`，同步字段：

- `ANTHROPIC_MODEL`
- `ANTHROPIC_SMALL_FAST_MODEL`
- `CLAUDE_CODE_MODEL`
- `CLAUDE_CODE_SMALL_MODEL`
- `CLAUDE_CODE_SUBAGENT_MODEL`
- `permissions.deny += Agent(Explore)`

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

目标命令不存在时，`ccc` 会在运行前检查并尝试自动安装。

## 升级

```bash
ccc upgrade --dry-run             # 预览
ccc upgrade                       # 执行升级
ccc upgrade --version 2.5.6       # 升级到指定版本
```

- Unix 平台支持原地升级
- Windows 建议重新执行 `install.ps1`

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
