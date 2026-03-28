# 配置指南

## 当前配置文件位置

新的 Go 实现默认使用：

```text
Linux / macOS: ~/.config/ccc/config.json
Windows: %APPDATA%\ccc\config.json
```

旧配置位置：

- `~/.ccc/config.json`
- `~/.cc-config.json`

仍然可以读取，但建议执行：

```bash
ccc config migrate
```

把内容写入新路径。

## 当前配置结构

当前配置文件是一个 JSON 对象，不再是旧版的数组格式。

示例：

```json
{
  "version": 1,
  "current_profile": "codex-relay",
  "profiles": [
    {
      "id": "codex-relay",
      "name": "Codex Relay",
      "provider": "custom",
      "command": "codex",
      "base_url": "https://relay.example.com",
      "api_key": "sk-xxx",
      "model": "gpt-5.4",
      "fast_model": "gpt-5.4-mini",
      "env": {
        "OPENAI_ORG": "example"
      },
      "sync_external": true
    }
  ]
}
```

## 字段说明

### 顶层字段

- `version`
  - 配置版本号
  - 当前固定为 `1`
- `current_profile`
  - 当前默认使用的 profile id
  - 可为空
- `profiles`
  - profile 列表

### profile 字段

- `id`
  - profile 唯一标识
  - 推荐使用小写字母、数字和 `-`
- `name`
  - 显示名称
- `provider`
  - provider 类型
  - 当前常见值：`anthropic`、`openai`、`zhipu`、`custom`、`legacy-import`
- `command`
  - 目标命令
  - 当前支持：`claude`、`codex`
- `base_url`
  - API 端点
- `api_key`
  - API key
- `model`
  - 主模型
- `fast_model`
  - 快速模型
  - 未设置时通常会退回主模型
- `env`
  - 额外环境变量
- `sync_external`
  - 是否在 `sync` / `run` 时同步外部配置文件

## Claude profile 示例

```json
{
  "id": "claude-official",
  "name": "Claude Official",
  "provider": "anthropic",
  "command": "claude",
  "base_url": "https://api.anthropic.com",
  "api_key": "sk-ant-xxx",
  "model": "claude-3-7-sonnet",
  "fast_model": "claude-3-5-haiku",
  "sync_external": true
}
```

运行时会注入：

- `ANTHROPIC_BASE_URL`
- `ANTHROPIC_AUTH_TOKEN`
- `ANTHROPIC_MODEL`
- `ANTHROPIC_SMALL_FAST_MODEL`
- `CLAUDE_CODE_MODEL`
- `CLAUDE_CODE_SMALL_MODEL`
- `CLAUDE_CODE_SUBAGENT_MODEL`

## Codex profile 示例

```json
{
  "id": "codex-relay",
  "name": "Codex Relay",
  "provider": "custom",
  "command": "codex",
  "base_url": "https://relay.example.com",
  "api_key": "sk-xxx",
  "model": "gpt-5.4",
  "fast_model": "gpt-5.4-mini",
  "sync_external": true
}
```

运行时会注入：

- `OPENAI_BASE_URL`
- `OPENAI_API_KEY`
- `OPENAI_MODEL`
- `OPENAI_SMALL_FAST_MODEL`

并同步：

- `~/.codex/config.toml`
- `~/.codex/auth.json`

`base_url` 在写入 Codex 配置文件时会规范化为 `/v1`。

## Zhipu Claude-compatible profile 示例

```json
{
  "id": "zhipu-claude-compatible",
  "name": "Zhipu Claude-Compatible",
  "provider": "zhipu",
  "command": "claude",
  "base_url": "https://open.bigmodel.cn/api/anthropic",
  "api_key": "sk-xxx",
  "model": "glm-5",
  "fast_model": "glm-4.7",
  "sync_external": true
}
```

## preset 辅助

当前 `profile add` 支持以下 preset：

- `anthropic`
- `openai`
- `zhipu`

例如：

```bash
ccc profile add --preset anthropic --api-key sk-ant-xxx
ccc profile add --preset openai --api-key sk-xxx
ccc profile add --preset zhipu --api-key sk-xxx
```

preset 会为以下字段提供默认值：

- `name`
- `provider`
- `command`
- `base_url`
- `model`
- `fast_model`

如果你显式传入这些字段，显式值优先。

## 从旧配置导入

旧 Bash 配置数组仍可被读取，例如：

```json
[
  {
    "name": "Claude (Official)",
    "env": {
      "ANTHROPIC_BASE_URL": "https://api.anthropic.com",
      "ANTHROPIC_AUTH_TOKEN": "token",
      "ANTHROPIC_MODEL": "claude-main"
    }
  }
]
```

当前 Go CLI 会在读取时自动把它转换为新结构，并标记：

- `provider = "legacy-import"`

但只有执行 `ccc config migrate` 后，才会真正写成新格式。

## 推荐操作方式

### 查看当前配置

```bash
ccc config show
```

### 查看配置文件路径

```bash
ccc config path
```

### 添加 profile

```bash
ccc profile add \
  --name "Codex Relay" \
  --preset openai \
  --base-url https://relay.example.com \
  --api-key sk-xxx \
  --model gpt-5.4
```

### 切换当前 profile

```bash
ccc profile use codex-relay
```

### 删除 profile

```bash
ccc profile delete codex-relay
```

### 更新 profile

```bash
ccc profile update codex-relay --model gpt-5.4
ccc profile update codex-relay --env OPENAI_ORG=example
ccc profile update codex-relay --unset-env OPENAI_ORG
ccc profile update codex-relay --clear-env --no-sync
```

## 安全建议

- 不要把配置文件提交到公开仓库
- `api_key` 当前仍以明文形式保存在本地配置文件中
- 如果共享机器，注意文件权限和 shell 历史

## 诊断

如果你不确定当前配置是否来自新路径、旧路径或是否已经迁移，优先执行：

```bash
ccc doctor
ccc config show
```
