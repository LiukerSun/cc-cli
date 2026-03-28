# CC-CLI

`ccc` 是一个用于管理 Claude / Codex 运行配置的命令行工具。

当前 `main` 分支正处于 Go 重构阶段。新的 Go CLI 已经具备以下能力：

- 基于 profile 的配置管理
- 新目录布局与旧配置迁移
- `claude` / `codex` 运行前依赖检查
- 缺失目标 CLI 时按需自动安装
- 同步 `~/.claude/settings.json`
- 同步 `~/.codex/config.toml` / `~/.codex/auth.json`

仓库中的 `bin/ccc` 和 `bin/ccc.ps1` 已降级为兼容 wrapper；新的 Go CLI 是当前唯一主实现。

## 安装

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/LiukerSun/cc-cli/main/install.sh | bash
```

### Windows PowerShell

```powershell
irm https://raw.githubusercontent.com/LiukerSun/cc-cli/main/install.ps1 | iex
```

### 安装行为说明

- 安装器优先安装 `ccc` 的预编译二进制
- 在本地仓库中执行安装脚本时，会直接用当前源码构建
- 安装 `ccc` 本体不再要求预先安装 Node.js
- `claude` / `codex` 的缺失检测与自动安装发生在运行 profile 时

默认安装路径：

- Linux / macOS: `~/.local/bin/ccc`
- Windows: `%LOCALAPPDATA%\\Programs\\ccc\\bin\\ccc.exe`

## 当前命令面

以 `ccc help` 输出为准：

```text
ccc help
ccc version
ccc current
ccc run [profile-id-or-name] [--dry-run] [--env-only] [-y|--bypass] [-- cli-args...]
ccc sync [profile-id-or-name] [--dry-run]
ccc profile list [--json]
ccc profile add [--name ...] [--preset anthropic|openai|zhipu] --api-key ...
ccc profile update <profile-id-or-name> [--preset anthropic|openai|zhipu] [--model ...]
ccc profile use <profile-id-or-name>
ccc profile delete <profile-id-or-name>
ccc paths [--json]
ccc config path
ccc config show
ccc config migrate
ccc doctor
ccc upgrade [--version <semver>] [--dry-run]
```

兼容别名当前只保留：

- `ccc --help`
- `ccc --version`

## 快速开始

### 1. 添加一个 profile

Anthropic:

```bash
ccc profile add \
  --preset anthropic \
  --api-key sk-ant-xxx \
  --name "Claude Official"
```

OpenAI:

```bash
ccc profile add \
  --preset openai \
  --api-key sk-xxx \
  --name "Codex OpenAI"
```

Zhipu:

```bash
ccc profile add \
  --preset zhipu \
  --api-key sk-xxx \
  --name "Zhipu Claude-Compatible"
```

自定义 relay:

```bash
ccc profile add \
  --name "Codex Relay" \
  --preset openai \
  --base-url https://relay.example.com \
  --api-key sk-xxx \
  --model gpt-5.4
```

### 2. 查看 profile

```bash
ccc profile list
ccc current
```

### 3. 切换当前 profile

```bash
ccc profile use codex-relay
```

### 4. 更新现有 profile

```bash
ccc profile update codex-relay --model gpt-5.4
ccc profile update codex-relay --base-url https://relay.example.com --env OPENAI_ORG=example
ccc profile update claude-official --id claude-prod --name "Claude Prod"
```

### 5. 预览运行计划

```bash
ccc run --dry-run -- --help
```

### 6. 同步外部配置

```bash
ccc sync
```

### 7. 运行目标 CLI

```bash
ccc run
ccc run codex-relay -- --help
```

### 8. 预览升级目标

```bash
ccc upgrade --dry-run
ccc upgrade --version 2.2.1 --dry-run
```

## 目录布局

当前 Go 实现使用新的目录模型：

### Linux / macOS

- Binary: `~/.local/bin/ccc`
- Config: `~/.config/ccc/config.json`
- Data: `~/.local/share/ccc/`
- Cache: `~/.cache/ccc/`
- State: `~/.local/state/ccc/`

### Windows

- Binary: `%LOCALAPPDATA%\\Programs\\ccc\\bin\\ccc.exe`
- Config: `%APPDATA%\\ccc\\config.json`
- Data: `%LOCALAPPDATA%\\ccc\\data\\`
- Cache: `%LOCALAPPDATA%\\ccc\\cache\\`
- State: `%LOCALAPPDATA%\\ccc\\state\\`

旧路径：

- `~/.ccc/config.json`
- `~/.cc-config.json`

仍可读取，并可通过 `ccc config migrate` 写入新路径。

## 运行时依赖

`ccc` 本体安装不依赖 Node.js，但目标 CLI 的自动安装依赖 Node.js / npm。

最低要求：

- `claude`: Node.js `>= 18.0.0`
- `codex`: Node.js `>= 16.0.0`

如果当前 profile 对应的 CLI 不存在，`ccc run` 会：

1. 检查 `node` 和 `npm`
2. 校验最低版本
3. 尝试执行 `npm install -g`
4. 成功后继续运行目标命令

## 升级

`ccc upgrade` 会根据当前平台解析对应的 GitHub Release 资产，并在 Unix 上尝试原地替换当前二进制。

示例：

```bash
ccc upgrade --dry-run
ccc upgrade
ccc upgrade --version 2.2.1
```

说明：

- Linux / macOS: 支持当前二进制的原地升级
- Windows: 当前只支持 `--dry-run` 规划；实际升级请重新执行 `install.ps1`

## 外部配置同步

### Claude

`ccc sync` 或 `ccc run` 会更新：

- `~/.claude/settings.json`

写入：

- `ANTHROPIC_MODEL`
- `ANTHROPIC_SMALL_FAST_MODEL`
- `CLAUDE_CODE_MODEL`
- `CLAUDE_CODE_SMALL_MODEL`
- `CLAUDE_CODE_SUBAGENT_MODEL`
- `permissions.deny += Agent(Explore)`

### Codex

`ccc sync` 或 `ccc run` 会更新：

- `~/.codex/config.toml`
- `~/.codex/auth.json`

写入：

- `model_provider = "codex"`
- `model = "<当前模型>"`
- `[model_providers.codex].base_url`
- `[model_providers.codex].wire_api = "responses"`
- `OPENAI_API_KEY`

## 开发

本地至少执行：

```bash
make test
make build
```

## 发布

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

重构说明见：

- [docs/go-refactor.md](/root/cc-cli/docs/go-refactor.md)

## 状态

当前仓库仍处于重构过渡期，后续预计继续推进：

- provider 辅助能力
- 更完整的命令兼容层
- 更薄的发布/安装体验
- 逐步淘汰旧 Bash / PowerShell 主逻辑
