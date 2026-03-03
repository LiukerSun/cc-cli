# CC-CLI

<div align="center">
  <strong>快速切换 AI 模型配置的命令行工具</strong>
</div>

<p align="center">
  交互式选择 • 一键切换 • 直接启动 Claude
</p>

<p align="center">
  <a href="https://github.com/LiukerSun/cc-cli/stargazers">
    <img src="https://img.shields.io/github/stars/LiukerSun/cc-cli.svg" alt="GitHub stars">
  </a>
  <a href="https://github.com/LiukerSun/cc-cli/issues">
    <img src="https://img.shields.io/github/issues/LiukerSun/cc-cli.svg" alt="GitHub issues">
  </a>
  <a href="https://github.com/LiukerSun/cc-cli/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/LiukerSun/cc-cli.svg" alt="License">
  </a>
</p>

---

## ✨ 特性

- 🎯 **交互式选择** - 使用上下键轻松切换模型
- 🚀 **直接启动** - 无需修改环境变量，直接启动 Claude
- 🔑 **API Key 管理** - 交互式添加、查看和编辑配置
- 🎨 **彩色输出** - 美观的终端界面
- ⚡ **零依赖** - 纯 Bash 实现，无外部依赖
- 🔄 **Bypass 模式** - 支持 `CLAUDE_SKIP_PERMISSIONS`
- 📦 **配置持久化** - 自动保存和恢复上次选择

## 📦 安装

### macOS / Linux (Bash)

```bash
curl -fsSL https://raw.githubusercontent.com/LiukerSun/cc-cli/main/install.sh | bash
```

### Windows (PowerShell)

```powershell
# 以管理员身份运行 PowerShell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser

# 一键安装
irm https://raw.githubusercontent.com/LiukerSun/cc-cli/main/install.ps1 | iex
```

或者手动安装：
```powershell
# 下载安装脚本
Invoke-WebRequest -Uri https://raw.githubusercontent.com/LiukerSun/cc-cli/main/install.ps1 -OutFile install.ps1

# 运行安装
.\install.ps1
```

### 手动安装

```bash
# 克隆仓库
git clone https://github.com/LiukerSun/cc-cli.git
cd cc-cli

# macOS/Linux
./install.sh

# Windows
.\install.ps1
```

### 验证安装

```bash
# macOS/Linux
cc --version
cc --help

# Windows PowerShell
cc -Version
cc -Help
```

## 🚀 快速开始

### 埥看所有模型

```bash
cc --list
```

### 交互式选择（推荐）

```bash
cc
```

使用 ↑↓ 选择模型，按 Enter 确认。

### 直接指定模型

```bash
cc 2    # 使用模型 #2
```

### 查看 API Keys

```bash
cc -s    # 显示所有配置的 API keys（部分隐藏）
```

### 编辑配置

```bash
cc -E    # 用 vim 打开配置文件
```

### 添加新模型

```bash
cc -a    # 交互式添加新模型
```

支持两种添加方式：
1. **厂商自动获取** - 选择 ZHIPU AI，输入 API Key，自动获取最新模型列表
2. **手动输入** - 手动填写所有配置信息

#### ZHIPU AI 快速配置示例

```bash
$ cc -a
Select provider:
  1) ZHIPU AI (智谱) - auto fetch models
  2) Manual input

Choice [1-2]: 1

API Key: your-zhipu-api-key

Fetching models from ZHIPU AI... Done!

Available Models:

   1) GLM-4-0520
   2) GLM-4-Air
   3) GLM-4-AirX
   ...

Select main model [1-N]: 2
Select fast model [1-N] (default: same as main): 3

✓ Model 'ZHIPU (GLM-4-Air)' added successfully!
```

## 📖 命令参考

### 基础命令
| 命令 | 说明 |
|------|------|
| `cc` | 交互式选择模型并启动 Claude |
| `cc -l, --list` | 列出所有可用模型 |
| `cc -c, --current` | 显示当前使用的模型 |
| `cc -h, --help` | 显示帮助信息 |

### 配置管理
| 命令 | 说明 |
|------|------|
| `cc -E, --edit` | 编辑配置文件 |
| `cc -a, --add` | 交互式添加新模型 |
| `cc -s, --show-keys` | 查看 API keys（部分隐藏） |

### 模型选择
| 命令 | 说明 |
|------|------|
| `cc 2` | 直接使用模型 #2 启动 Claude |
| `cc -y 3` | Bypass 模式 + 模型 #3 |
| `cc 1 -- --help` | 传递参数给 Claude |

### 环境变量
| 命令 | 说明 |
|------|------|
| `cc -e 2` | 仅设置环境变量，不启动 Claude |

## ⚙ 交互式选择

运行 `cc` 会显示交互式菜单：

```
═══════════════════════════════════════
  Available AI Models
═══════════════════════════════════════

    1) ZHIPU AI
  ➜ 2) MiniMax (China)        ← 当前选中（蓝色高亮）
    3) Kimi (Moonshot AI)
  ➜ 4) claude (current)       ← 上次使用（绿色标记）

───────────────────────────────────────

  ↑↓ Navigate  •  Enter Select  •  q Exit
```

### 快捷键
- `↑` / `↓` - 上下选择
- `Enter` - 确认并启动
- `1-9` - 快速选择
- `q` / `ESC` - 退出

## ⚙️ 配置

### 配置文件位置

```
~/.cc-config.json
```

### 配置格式

```json
[
    {
        "name": "模型名称",
        "env": {
            "ANTHROPIC_BASE_URL": "API地址",
            "ANTHROPIC_AUTH_TOKEN": "API密钥",
            "ANTHROPIC_MODEL": "主模型",
            "ANTHROPIC_SMALL_FAST_MODEL": "快速模型"
        }
    }
]
```

### 示例配置

```json
[
    {
        "name": "Claude (Official)",
        "env": {
            "ANTHROPIC_BASE_URL": "https://api.anthropic.com",
            "ANTHROPIC_AUTH_TOKEN": "sk-ant-xxxxx",
            "ANTHROPIC_MODEL": "claude-3-5-sonnet-20241022",
            "ANTHROPIC_SMALL_FAST_MODEL": "claude-3-5-haiku-20241022"
        }
    },
    {
        "name": "OpenAI GPT-4",
        "env": {
            "ANTHROPIC_BASE_URL": "https://api.openai.com/v1",
            "ANTHROPIC_AUTH_TOKEN": "sk-xxxxx",
            "ANTHROPIC_MODEL": "gpt-4o",
            "ANTHROPIC_SMALL_FAST_MODEL": "gpt-4o-mini"
        }
    }
]
```

## 🔧 高级用法

### 自定义编辑器

```bash
# 使用 VS Code
EDITOR=code cc -E

# 使用 nano
EDITOR=nano cc -E

# 永久修改
echo 'export EDITOR=code' >> ~/.zshrc
source ~/.zshrc
```

### Bypass Permissions

```bash
# 启动并启用 bypass
cc -y 2

# 或
cc --bypass 2
```

### 传递参数给 Claude

```bash
# 传递 --help
cc 1 -- --help

# 传递提示词
cc 2 -- "Write a hello world program"

# 传递多个参数
cc 3 -- --version --verbose
```

### 仅设置环境变量

```bash
# 仅设置环境变量
cc -e 2

# 然后手动启动
source /tmp/cc-model-env.sh
claude
```

## 🗑️ 卸载

```bash
./install.sh --uninstall
```

## 🤝 贡献

欢迎贡献！请查看 [Contributing Guide](CONTRIBUTING.md)

## 📝 更新日志

查看 [CHANGELOG.md](CHANGELOG.md)

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件

## 🙏 致谢

- 感谢 [Anthropic](https://anthropic.com) 提供 Claude API
- 所有贡献者

## 📮 问题反馈

- 提交 Issue: https://github.com/LiukerSun/cc-cli/issues
- 功能请求: https://github.com/LiukerSun/cc-cli/issues/new

## 🔗 相关项目

- [Claude CLI](https://claude.ai) - Claude 官方命令行工具
- [Anthropic API](https://docs.anthropic.com) - Anthropic API 文档

---

<div align="center">
  Made with ❤️ by the community
</div>
