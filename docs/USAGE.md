# 使用指南

## 当前命令

当前 Go CLI 支持：

```bash
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

## 查看版本和路径

```bash
ccc version
ccc paths
ccc config path
```

如果需要机器可读输出：

```bash
ccc paths --json
ccc profile list --json
ccc config show
```

## Profile 管理

### 使用 Anthropic preset

```bash
ccc profile add \
  --preset anthropic \
  --api-key sk-ant-xxx \
  --name "Claude Official"
```

### 使用 OpenAI preset

```bash
ccc profile add \
  --preset openai \
  --api-key sk-xxx \
  --name "Codex OpenAI"
```

### 使用 Zhipu preset

```bash
ccc profile add \
  --preset zhipu \
  --api-key sk-xxx \
  --name "Zhipu Claude-Compatible"
```

### 在 preset 基础上覆盖默认值

```bash
ccc profile add \
  --name "Codex Relay" \
  --preset openai \
  --base-url https://relay.example.com \
  --api-key sk-xxx \
  --model gpt-5.4
```

### 列出 profile

```bash
ccc profile list
```

### 选择当前 profile

```bash
ccc profile use codex-relay
ccc current
```

### 删除 profile

```bash
ccc profile delete codex-relay
```

### 更新已有 profile

```bash
ccc profile update codex-relay --model gpt-5.4
ccc profile update codex-relay --base-url https://relay.example.com --env OPENAI_ORG=example
ccc profile update claude-official --id claude-prod --name "Claude Prod"
```

可用更新能力包括：

- 切换 preset
- 覆盖 `name`、`id`、`command`、`provider`
- 覆盖 `base_url`、`api_key`、`model`、`fast_model`
- `--env KEY=VALUE` 追加或覆盖额外环境变量
- `--unset-env KEY` 删除额外环境变量
- `--clear-env` 清空额外环境变量
- `--sync` / `--no-sync` 调整外部同步行为

## 运行目标 CLI

### 直接运行当前 profile

```bash
ccc run
```

### 指定 profile 运行

```bash
ccc run codex-relay
```

### 透传参数给目标 CLI

```bash
ccc run -- --help
ccc run codex-relay -- --version
ccc run codex-relay -- "Write a hello world program"
```

### Dry run

```bash
ccc run --dry-run
```

会显示：

- 选中的 profile
- 将执行的命令
- 透传参数
- 环境变量
- 将同步的外部配置路径

### 只打印环境变量

```bash
ccc run --env-only
```

### Bypass 模式

```bash
ccc run --bypass
ccc run -y
```

映射规则：

- `claude`: `--dangerously-skip-permissions`
- `codex`: `--dangerously-bypass-approvals-and-sandbox`

## 外部配置同步

### 预览同步目标

```bash
ccc sync --dry-run
```

### 执行同步

```bash
ccc sync
ccc sync codex-relay
```

同步规则：

- `claude` profile 写 `~/.claude/settings.json`
- `codex` profile 写 `~/.codex/config.toml` 和 `~/.codex/auth.json`

`ccc run` 在真正执行目标命令前，也会自动触发同步。

## 配置迁移

如果你还在使用旧配置：

- `~/.ccc/config.json`
- `~/.cc-config.json`

可以先查看当前读取来源：

```bash
ccc config show
```

然后执行：

```bash
ccc config migrate
```

把内容写入新路径。

## 依赖检查和自动安装

### 检查环境

```bash
ccc doctor
```

当前会检查：

- 目录布局
- 配置来源
- profile 数量
- 旧路径残留
- `node`
- `npm`
- `claude`
- `codex`
- 当前 profile 命令是否在 PATH

### 自动安装目标 CLI

如果当前 profile 对应的命令不存在，`ccc run` 会尝试自动安装。

要求：

- Claude CLI: Node.js `>= 18.0.0`
- Codex CLI: Node.js `>= 16.0.0`

自动安装使用：

```bash
npm install -g @anthropic-ai/claude-code
npm install -g @openai/codex
```

## 升级 `ccc`

### 预览升级计划

```bash
ccc upgrade --dry-run
ccc upgrade --version 2.2.1 --dry-run
```

输出会包含：

- 当前版本
- 目标版本
- 对应平台的 release 资产名
- 下载地址和校验文件地址
- 当前二进制路径

### 执行升级

```bash
ccc upgrade
ccc upgrade --version 2.2.1
```

当前行为：

- Linux / macOS: 尝试下载并原地替换当前 `ccc` 二进制
- Windows: 当前仅支持 dry-run 规划，实际升级请重新运行 `install.ps1`

## 常见问题

### 1. `ccc doctor` 提示当前命令不在 PATH

说明当前 profile 指向的 `claude` 或 `codex` 还不可执行。先确认：

```bash
node --version
npm --version
```

然后重新执行：

```bash
ccc run --dry-run
ccc run
```

### 2. 旧配置还没有迁移

先检查来源：

```bash
ccc config show
```

如果来源仍然是 `legacy-root` 或 `legacy-file`，执行：

```bash
ccc config migrate
```

### 3. 安装后找不到 `ccc`

确认安装位置：

```bash
ccc paths
```

如果提示 binary dir 不在 PATH，把它加入 shell 配置。
