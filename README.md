# CC-CLI

<div align="center">
  <strong>快速切换 AI 模型配置的命令行工具</strong>
</div>

<p align="center">
  交互式选择 • 一键切换 • 直接启动 Claude / Codex
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
- 🚀 **直接启动** - 无需手动切换配置，直接启动 Claude 或 Codex
- 🛠️ **CLI 自动补装** - 检测缺失的 Claude / Codex CLI，并在 Node.js / npm 可用时自动通过 npm 安装
- 🔑 **API Key 管理** - 交互式添加、查看和编辑配置
- 🎨 **彩色输出** - 美观的终端界面
- ⚡ **轻依赖** - 核心 Bash 脚本无额外依赖；安装器和 CLI 自动补装依赖 Node.js / npm
- 🔄 **Bypass 模式** - 支持 `CLAUDE_SKIP_PERMISSIONS`
- 📦 **配置持久化** - 自动保存和恢复上次选择
- 🗂️ **统一目录布局** - `ccc` 自身配置和安装文件统一收拢到 `~/.ccc/`
- 🤖 **Team / Codex 配置同步** - 启动 Claude 时自动同步 `~/.claude/settings.json`；启动 Codex 时自动同步 `~/.codex/config.toml` 和 `~/.codex/auth.json`
- 🔀 **多提供商支持** - 支持 Anthropic 兼容提供商、智谱 AI、阿里云百炼（Coding Plan）和 OpenAI Codex
- 📡 **自动获取模型** - Claude-compatible provider 支持从 API 获取模型列表；Codex 支持内置官方模型列表

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

> 安装器要求系统已安装 Node.js；如果未安装，会直接退出并提示先安装 Node.js。
> 安装器会自动检测 `claude` 和 `codex` 是否已安装。
> 缺失时会尽力执行 `npm install -g` 自动补装，但不会因为未使用的 CLI 缺失而中断 `ccc` 本身的安装。
> 如果缺少 Node.js / npm，或 Node.js 版本过低，会明确提示当前版本和最低要求；首次实际运行对应命令时，`ccc` 也会再次按需尝试安装。
> 当前最低要求为：`claude` 需要 Node.js `>= 18.0.0`，`codex` 需要 Node.js `>= 16.0.0`。

### 验证安装

```bash
# macOS/Linux
ccc --version
ccc --help

# Windows PowerShell
ccc --version
ccc --help
```

> 💡 Windows PowerShell 不区分大小写，`-V` 和 `-v`、`-U` 和 `-u` 等效果相同。
> 为避免冲突，部分命令仅支持长选项（如 `--edit`、`--validate`、`--uninstall`）。

## 🚀 快速开始

### 查看所有模型

```bash
ccc --list
```

### 交互式选择（推荐）

```bash
ccc
```

使用 ↑↓ 选择模型，按 Enter 确认。

### 直接指定模型

```bash
ccc 2    # 使用模型 #2
```

### 查看 API Keys

```bash
ccc -s    # 显示所有配置的 API keys（部分隐藏）
```

### 编辑配置

```bash
ccc --edit    # 用 vim 打开配置文件
```

### 添加新模型

```bash
ccc -a    # 交互式添加新模型
```

支持四种添加方式：
1. **ZHIPU AI 自动获取** - 选择 ZHIPU AI，输入 API Key，自动获取最新模型列表
2. **Alibaba Coding Plan (百炼)** - 选择阿里云百炼，输入 API Key，自动获取编程模型列表
3. **OpenAI Codex** - 输入 API Base URL / API Key，从内置官方 Codex 模型列表中选择
4. **手动输入（Claude-compatible）** - 手动填写 Claude-compatible 配置信息

#### ZHIPU AI 快速配置示例

```bash
$ ccc -a
Select provider:
  1) ZHIPU AI - Claude-compatible provider
  2) Alibaba Coding Plan - Claude-compatible provider
  3) OpenAI Codex
  4) Manual input (Claude-compatible)

Choice [1-4]: 1

API Key: your-zhipu-api-key

Fetching models from ZHIPU AI... Done!

Available Models:

   1) GLM-4-0520
   2) GLM-4-Air
   3) GLM-4-AirX
   ...

Select main model [1-N]: 2
Select fast model [1-N] (default: same as main): 3

Model 'ZHIPU (GLM-4-Air)' added successfully!
```

#### Alibaba Coding Plan (百炼) 快速配置示例

```bash
$ ccc -a
Select provider:
  1) ZHIPU AI - Claude-compatible provider
  2) Alibaba Coding Plan - Claude-compatible provider
  3) OpenAI Codex
  4) Manual input (Claude-compatible)

Choice [1-4]: 2

API Key: your-dashscope-api-key

Fetching models from Alibaba Coding Plan... Done!

Available Models:

   1) qwen3.5-plus
   2) qwen3-max-2026-01-23
   3) qwen3-coder-next
   4) qwen3-coder-plus
   5) glm-5
   6) glm-4.7
   7) kimi-k2.5
   8) minimax-m2.5

Select main model [1-8]: 1
Select fast model [1-8] (default: same as main): 4

Model 'Alibaba Coding Plan (qwen3.5-plus)' added successfully!
```

#### OpenAI Codex 快速配置示例

```bash
$ ccc -a
Select provider:
  1) ZHIPU AI - Claude-compatible provider
  2) Alibaba Coding Plan - Claude-compatible provider
  3) OpenAI Codex
  4) Manual input (Claude-compatible)

Choice [1-4]: 3

API Base URL (e.g., https://api.openai.com/v1): https://relay.example.com
API Key: sk-xxx

Using built-in OpenAI/Codex model list

Available Models:

   1) gpt-5.4
   2) gpt-5.4-mini
   3) gpt-5.3-codex
   ...

Select model [1-N]: 1

Model 'Codex (gpt-5.4)' added successfully!
```

## 📖 命令参考

### 基础命令
| 命令 | 说明 |
|------|------|
| `ccc` | 交互式选择模型并启动对应 CLI |
| `ccc -l, --list` | 列出所有可用模型 |
| `ccc -c, --current` | 显示当前使用的模型 |
| `ccc -V, --version` | 显示版本号 |
| `ccc -h, --help` | 显示帮助信息 |

### 配置管理
| 命令 | 说明 |
|------|------|
| `ccc --edit` | 编辑配置文件 |
| `ccc -a, --add` | 交互式添加新模型 |
| `ccc -d, --delete N` | 删除模型 #N |
| `ccc -s, --show` | 查看 API keys（部分隐藏） |
| `ccc --validate` | 验证并修复配置文件 |
| `ccc -U, --upgrade` | 升级到最新版本 |
| `ccc --uninstall` | 卸载 cc-cli |

### 模型选择
| 命令 | 说明 |
|------|------|
| `ccc 2` | 直接使用模型 #2 启动对应 CLI |
| `ccc -y 3` | Bypass 模式 + 模型 #3 |
| `ccc 1 -- --help` | 传递参数给目标 CLI |

### 环境变量
| 命令 | 说明 |
|------|------|
| `ccc -e 2` | 仅设置环境变量，不启动目标 CLI |

## ⚙ 交互式选择

运行 `ccc` 会显示交互式菜单：

```
═══════════════════════════════════════
  Available AI Models
═══════════════════════════════════════

    1) ZHIPU AI
  ➜ 2) MiniMax (China)        ← 当前选中（蓝色高亮）
    3) Kimi (Moonshot AI)
    4) Codex (gpt-5.4)        ← 其他已保存模型

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
~/.ccc/config.json
```

升级到新版本后，如果检测到旧的 `~/.cc-config.json` / `~/.cc-cli/`，`ccc` 会自动迁移到 `~/.ccc/`。

### 配置格式

```json
[
    {
        "name": "模型名称",
        "command": "可选，默认 claude；使用 Codex 时填 codex",
        "env": {
            "ANTHROPIC_BASE_URL": "API地址",
            "ANTHROPIC_AUTH_TOKEN": "API密钥",
            "ANTHROPIC_MODEL": "主模型",
            "ANTHROPIC_SMALL_FAST_MODEL": "快速模型"
        }
    }
]
```

`ccc -a` 的 `codex` 分支会使用内置的 OpenAI/Codex 模型列表供选择，不请求远端 `/models`；如果需要特殊模型，也可以手动输入自定义模型 ID。
除 `OpenAI Codex` 选项外，其余新增入口默认都是 Claude-compatible 配置。

### Codex 配置同步

当你选择 `command = "codex"` 的配置启动时，`ccc` 会自动同步：
- `~/.codex/config.toml`
  - `model_provider = "codex"`
  - `model = "<当前模型>"`
  - `[model_providers.codex].base_url = "<你的 OpenAI 兼容 /v1 地址>"`
  - `wire_api = "responses"`
- `~/.codex/auth.json`
  - `OPENAI_API_KEY`

如果你在配置中填的是根域名，例如 `https://relay.example.com`，`ccc` 会自动规范化成 `https://relay.example.com/v1` 再写入 Codex 配置。

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
        "command": "codex",
        "env": {
            "OPENAI_BASE_URL": "https://api.openai.com/v1",
            "OPENAI_API_KEY": "sk-xxxxx",
            "OPENAI_MODEL": "gpt-4o"
        }
    },
    {
        "name": "Codex (API易)",
        "command": "codex",
        "env": {
            "OPENAI_BASE_URL": "https://vip.apiyi.com/v1",
            "OPENAI_API_KEY": "sk-xxxxx",
            "OPENAI_MODEL": "o3"
        }
    }
]
```

## 🔧 高级用法

### Claude Team Subagent 模型配置

当你使用 Claude Code 的 team 功能，且在 `ccc` 中选择的是 `claude` 配置时，subagent 会自动使用你在脚本中选择的模型。

```bash
# 选择模型后启动 Claude
ccc 1  # 选择 qwen3.5-plus

# 在 Claude Code 中创建 team
/team 创建团队任务
```

此时创建的 subagent 会使用 `qwen3.5-plus` 模型，而不是默认的 `haiku` 或 `claude-opus-4-6`。

**工作原理：**
- `ccc` 脚本会自动更新 `~/.claude/settings.json` 文件
- 将当前模型的 `ANTHROPIC_MODEL` 和 `ANTHROPIC_SMALL_FAST_MODEL` 写入配置
- 同时设置 `CLAUDE_CODE_MODEL`、`CLAUDE_CODE_SMALL_MODEL`、`CLAUDE_CODE_SUBAGENT_MODEL` 环境变量
- Claude Code 启动时读取这些环境变量，subagent 也会继承
- 自动禁用 Explore subagent（因为它硬编码使用 Haiku 模型，自定义 API 提供商不支持）

**环境变量说明：**
| 环境变量 | 说明 |
|----------|------|
| `CLAUDE_CODE_MODEL` | 主模型配置，覆盖所有 Agent 的默认模型 |
| `CLAUDE_CODE_SMALL_MODEL` | 快速模型配置，用于简单任务 |
| `CLAUDE_CODE_SUBAGENT_MODEL` | 专门用于 Team Subagent 的模型配置 |

### 升级到最新版本

```bash
# 检查并升级到最新版本
ccc -U

# 或
ccc --upgrade
```

升级功能会：
- 自动检查 GitHub 上的最新版本
- 如果有新版本，自动下载并安装
- 保留你的配置文件和 API keys

### 自定义编辑器

```bash
# 使用 VS Code
EDITOR=code ccc --edit

# 使用 nano
EDITOR=nano ccc --edit

# 永久修改
echo 'export EDITOR=code' >> ~/.zshrc
source ~/.zshrc
```

### Bypass Permissions

```bash
# 启动并启用 bypass
ccc -y 2

# 或
ccc --bypass 2
```

### 传递参数给目标 CLI

```bash
# 传递 --help
ccc 1 -- --help

# 传递提示词
ccc 2 -- "Write a hello world program"

# 传递多个参数
ccc 3 -- --version --verbose
```

### 仅设置环境变量

```bash
# 仅设置环境变量
ccc -e 2

# 然后手动启动
source ~/.ccc/tmp/cc-model-env.sh
claude
```

## 🔧 故障排除

### 常见问题

**1. Claude / Codex 未安装或 Node.js 版本过低**
```bash
# install.sh / install.ps1 要求机器上先有 Node.js
# ccc / install.sh / install.ps1 会优先自动检测并尝试安装缺失 CLI
# 在 Node.js 已安装的前提下，安装器阶段对未使用 CLI 仍是 best-effort

# 如需手动安装：
npm install -g @anthropic-ai/claude-code
npm install -g @openai/codex

# 如果提示 Node.js 版本过低：
# claude 需要 Node.js >= 18.0.0
# codex 需要 Node.js >= 16.0.0

# 如果运行时提示 "Cannot install claude/codex CLI automatically"
# 表示 ccc 已安装，但当前机器还不满足目标 CLI 的运行依赖
#
# 如果安装器提示机器没有 Node.js：
# macOS / Linux 推荐优先使用 nvm 安装
# curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.4/install.sh | bash
```

**2. 配置文件损坏**
```bash
# 备份并重新创建
cp ~/.ccc/config.json ~/.ccc/config.json.backup
ccc --add
```

**3. 权限问题**
```bash
# 确保安装目录权限
chmod 755 ~/.ccc
chmod +x ~/bin/ccc
```

**4. Subagent 不使用自定义模型**
- 确保运行 `ccc` 命令选择模型后启动 Claude
- `ccc` 会自动更新 `~/.claude/settings.json`
- 检查 `~/.claude/settings.json` 中是否有正确的环境变量配置

**5. API 请求失败**
- 检查 API Key 是否正确
- 确认 API 端点 URL 正确
- 查看网络防火墙设置

### Windows 特定问题

Windows PowerShell 用户请参考 [Windows 故障排除指南](docs/windows-troubleshooting.md)

## 🗑️ 卸载

### macOS / Linux

```bash
# 使用安装脚本
./install.sh --uninstall

# 或使用 ccc 命令
ccc --uninstall
```

### Windows

```powershell
# 使用安装脚本
.\install.ps1 -Action uninstall

# 或使用 ccc 命令
ccc --uninstall
```

卸载时会提示确认，并可选择：
- 是否删除配置文件
- 是否清理 Claude settings 文件

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
