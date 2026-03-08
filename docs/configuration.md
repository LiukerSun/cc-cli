# 配置指南

## 配置文件位置

CC-CLI 的配置文件位于 `~/.cc-config.json`

## 配置格式

配置文件是一个 JSON 数组，每个对象代表一个 AI 模型配置。

```json
[
    {
        "name": "模型显示名称",
        "env": {
            "ANTHROPIC_BASE_URL": "API 端点 URL",
            "ANTHROPIC_AUTH_TOKEN": "API 密钥",
            "ANTHROPIC_MODEL": "主模型 ID",
            "ANTHROPIC_SMALL_FAST_MODEL": "快速模型 ID"
        }
    }
]
```

## 必需字段

- `name` - 模型的显示名称（必填）
- `env.ANTHROPIC_BASE_URL` - API 端点 URL（必填）
- `env.ANTHROPIC_AUTH_TOKEN` - API 密钥（必填）
- `env.ANTHROPIC_MODEL` - 主模型 ID（必填）
- `env.ANTHROPIC_SMALL_FAST_MODEL` - 快速模型 ID（可选，默认与主模型相同)

## 添加新模型

### 方法 1：交互式添加（推荐）

```bash
cc --add
```

按提示输入：
- 模型名称
- API 端点 URL
- API 密钥
- 主模型 ID
- 快速模型 ID（可选）

### 方法 2：编辑配置文件

```bash
cc --edit
```

在 JSON 数组中添加新对象。

### 方法 3：手动编辑

```bash
nano ~/.cc-config.json
# 或
vim ~/.cc-config.json
```

## 示例配置

### Claude 官方 API

```json
{
    "name": "Claude (Official)",
    "env": {
        "ANTHROPIC_BASE_URL": "https://api.anthropic.com",
        "ANTHROPIC_AUTH_TOKEN": "sk-ant-api03-xxxxx",
        "ANTHROPIC_MODEL": "claude-3-5-sonnet-20241022",
        "ANTHROPIC_SMALL_FAST_MODEL": "claude-3-5-haiku-20241022"
    }
}
```

### OpenAI 端点

```json
{
    "name": "OpenAI GPT-4",
    "env": {
        "ANTHROPIC_BASE_URL": "https://api.openai.com/v1",
        "ANTHROPIC_AUTH_TOKEN": "sk-xxxxx",
        "ANTHROPIC_MODEL": "gpt-4o",
        "ANTHROPIC_SMALL_FAST_MODEL": "gpt-4o-mini"
    }
}
```

### 国产模型

```json
{
    "name": "智谱 AI",
    "env": {
        "ANTHROPIC_BASE_URL": "https://open.bigmodel.cn/api/anthropic",
        "ANTHROPIC_AUTH_TOKEN": "your-zhipu-api-key",
        "ANTHROPIC_MODEL": "glm-4.7",
        "ANTHROPIC_SMALL_FAST_MODEL": "glm-4.0"
    }
},
{
    "name": "Kimi (Moonshot AI)",
    "env": {
        "ANTHROPIC_BASE_URL": "https://api.moonshot.cn/anthropic/",
        "ANTHROPIC_AUTH_TOKEN": "your-kimi-api-key",
        "ANTHROPIC_MODEL": "kimi-k2.5",
        "ANTHROPIC_SMALL_FAST_MODEL": "kimi-k2-thinking"
    }
}
```

## API Key 安全

- 配置文件包含敏感信息，请妥善保管
- 不要将配置文件提交到公开仓库
- API Key 在 `--show-keys` 输出中会部分隐藏（显示前8位和后4位）
- 使用环境变量 `EDITOR` 可以指定编辑器

- 巻加 `.cc-config.json` 到 `.gitignore`

## 查看配置

```bash
# 查看所有模型
cc --list

# 查看 API keys（部分隐藏)
cc --show-keys

# 查看当前模型
cc --current
```

## 故障排除

### 配置文件损坏
```bash
# 备份并重新创建
cp ~/.cc-config.json ~/.cc-config.json.backup
cc --add
```

### 权限问题
```bash
# 确保安装目录权限
chmod 755 ~/.cc-cli
chmod +x ~/bin/cc
```

### Claude 未安装
如果 Claude CLI 未安装，cc 仍然可以管理配置，但无法启动 Claude。

安装 Claude: https://claude.ai

## 灵活配置
- 可以添加任意兼容 Anthropic API 的端点
- 可以配置多个同一提供商的不同模型
- 支持 custom headers（通过环境变量）

## Team Subagent 模型配置

当你使用 Claude Code 的 team 功能时，subagent 默认会使用硬编码的模型（如 `haiku` 或 `claude-opus-4-6`）。

从 v1.3.3 开始，`cc` 脚本会在启动时自动更新 `~/.claude/settings.json` 文件，将当前选择的模型配置写入其中，这样 team subagent 也会使用相同的模型。

### 工作原理

1. 当你运行 `cc [模型索引]` 时
2. 脚本会读取该模型的 `ANTHROPIC_MODEL` 和 `ANTHROPIC_SMALL_FAST_MODEL` 配置
3. 自动更新 `~/.claude/settings.json` 的 `env` 字段
4. Claude Code 启动时会读取这些环境变量，subagent 也会继承

### 示例

运行以下命令后：
```bash
cc 1  # 选择阿里百炼 qwen3.5-plus
```

`~/.claude/settings.json` 会被更新为：
```json
{
  "env": {
    "ANTHROPIC_MODEL": "qwen3.5-plus",
    "ANTHROPIC_SMALL_FAST_MODEL": "qwen3.5-plus",
    ...
  },
  ...
}
```

### 注意事项

- 如果你手动编辑了 `~/.claude/settings.json`，下次运行 `cc` 时会被覆盖
- 如果需要在不同项目中使用不同模型，可以在项目根目录创建 `.claude/settings.local.json`
- 此功能需要 `jq` 命令支持（macOS 和 Linux 默认已安装）
