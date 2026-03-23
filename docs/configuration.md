# 配置指南

## 配置文件位置

CC-CLI 的配置文件位于：

```text
~/.cc-config.json
```

配置文件是一个 JSON 数组；每个对象代表一个可选模型。

## 配置结构

```json
[
  {
    "name": "模型显示名称",
    "command": "可选，默认 claude；使用 Codex 时填 codex",
    "env": {
      "ANTHROPIC_BASE_URL": "API 端点 URL",
      "ANTHROPIC_AUTH_TOKEN": "API 密钥",
      "ANTHROPIC_MODEL": "主模型 ID",
      "ANTHROPIC_SMALL_FAST_MODEL": "快速模型 ID"
    }
  }
]
```

## 字段说明

- `name`
  - 模型显示名称，必填
- `command`
  - 启动命令，可选
  - 默认值为 `claude`
  - 目前支持：`claude`、`codex`
- `env`
  - 目标命令需要的环境变量集合，必填

### `command = "claude"` 时的必需字段

- `env.ANTHROPIC_BASE_URL`
- `env.ANTHROPIC_AUTH_TOKEN`
- `env.ANTHROPIC_MODEL`

可选字段：

- `env.ANTHROPIC_SMALL_FAST_MODEL`
  - 未设置时默认与主模型相同

### `command = "codex"` 时的必需字段

- `env.OPENAI_BASE_URL`
- `env.OPENAI_API_KEY`
- `env.OPENAI_MODEL`

可选字段：

- `env.OPENAI_SMALL_FAST_MODEL`
  - 未设置时默认与主模型相同

## 推荐配置方式

### 方法 1：交互式添加

```bash
ccc --add
```

当前交互式入口包括：

1. ZHIPU AI
2. Alibaba Coding Plan
3. OpenAI Codex
4. Manual input (Claude-compatible)

其中：
- `OpenAI Codex` 分支使用内置的 Codex 模型列表，不请求远端 `/models`
- 其他入口默认生成 Claude-compatible 配置

### 方法 2：编辑配置文件

```bash
ccc --edit
```

### 方法 3：手动编辑

```bash
nano ~/.cc-config.json
```

## 示例配置

### Claude-compatible

```json
{
  "name": "Claude (Official)",
  "env": {
    "ANTHROPIC_BASE_URL": "https://api.anthropic.com",
    "ANTHROPIC_AUTH_TOKEN": "sk-ant-xxxxx",
    "ANTHROPIC_MODEL": "claude-3-5-sonnet-20241022",
    "ANTHROPIC_SMALL_FAST_MODEL": "claude-3-5-haiku-20241022"
  }
}
```

### OpenAI / Codex

```json
{
  "name": "Codex (OpenAI)",
  "command": "codex",
  "env": {
    "OPENAI_BASE_URL": "https://api.openai.com/v1",
    "OPENAI_API_KEY": "sk-xxxxx",
    "OPENAI_MODEL": "gpt-5.4"
  }
}
```

### 第三方 OpenAI 兼容中转

```json
{
  "name": "Codex (Relay)",
  "command": "codex",
  "env": {
    "OPENAI_BASE_URL": "https://relay.example.com",
    "OPENAI_API_KEY": "sk-xxxxx",
    "OPENAI_MODEL": "gpt-5.4-mini"
  }
}
```

说明：
- 对于 Codex 配置，`ccc` 会在运行时把根域名自动规范化为 `/v1`
- 例如 `https://relay.example.com` 会写入为 `https://relay.example.com/v1`

## Codex 配置同步

当选中的模型 `command = "codex"` 时，`ccc` 会自动同步：

- `~/.codex/config.toml`
  - `model_provider = "codex"`
  - `model = "<当前模型>"`
  - `[model_providers.codex].base_url = "<规范化后的 /v1 地址>"`
  - `wire_api = "responses"`
- `~/.codex/auth.json`
  - `OPENAI_API_KEY`

注意：
- `ccc` 不会在 Codex 路径下继续注入旧的 `OPENAI_BASE_URL` 运行时环境变量
- 相关信息会优先写入 Codex 官方配置文件

## Claude Team Subagent 同步

当选中的模型 `command = "claude"` 时，`ccc` 会自动更新：

```text
~/.claude/settings.json
```

写入内容包括：

- `CLAUDE_CODE_MODEL`
- `CLAUDE_CODE_SMALL_MODEL`
- `CLAUDE_CODE_SUBAGENT_MODEL`

这样 Claude Code 的 team subagent 会跟随当前选择的模型。

## 查看和校验配置

```bash
ccc --list
ccc --show
ccc --current
ccc --validate
```

说明：
- `--show` 只会部分显示 API Key
- `--validate` 会根据 `command` 动态检查 `ANTHROPIC_*` 或 `OPENAI_*` 字段
- 在安装了 `jq` 的环境中，`--validate` 可以更可靠地修复无效项

## 安全建议

- 不要把 `~/.cc-config.json` 提交到公开仓库
- 建议把 `.cc-config.json` 加入你的项目 `.gitignore`
- 如果需要使用 GUI 编辑器，可以设置环境变量 `EDITOR`

例如：

```bash
EDITOR=code ccc --edit
```

## 与 CLI 安装相关的行为

- `install.sh` / `install.ps1` 运行前要求机器已安装 Node.js
- 在 Node.js 已安装的前提下，安装器会尽力自动补装缺失的 `claude` / `codex`
- 运行 `ccc` 时，如果当前目标 CLI 缺失，脚本也会按需再次尝试安装

手动安装命令：

```bash
npm install -g @anthropic-ai/claude-code
npm install -g @openai/codex
```

最低 Node.js 要求：

- Claude CLI: `>= 18.0.0`
- Codex CLI: `>= 16.0.0`

## 常见问题

### 配置文件损坏

```bash
cp ~/.cc-config.json ~/.cc-config.json.backup
ccc --add
```

### 权限问题

```bash
chmod 755 ~/.cc-cli
chmod +x ~/bin/ccc
```

### `--validate` 修复能力有限

如果当前环境没有 `jq`，Bash 版本只能做基础检查；复杂损坏场景建议直接：

```bash
ccc --edit
```

或重新创建配置。
