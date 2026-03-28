# Windows 故障排除

本页主要覆盖 PowerShell 环境下最常见的三类问题：

1. 安装器直接退出
2. 配置文件无法解析
3. 运行时找不到 `claude` / `codex`

## 1. 安装器直接退出

### 提示：`Node.js is required to install ccc`

这表示你在运行：

```powershell
irm https://raw.githubusercontent.com/LiukerSun/cc-cli/main/install.ps1 | iex
```

时，系统里还没有可用的 Node.js。

处理方法：

1. 安装 Node.js
2. 重新打开 PowerShell
3. 再次执行安装脚本

安装完成后可先确认：

```powershell
node --version
npm --version
```

### 最低版本要求

- Claude CLI: Node.js `>= 18.0.0`
- Codex CLI: Node.js `>= 16.0.0`

说明：
- 安装器要求机器上先有 Node.js
- 在 Node.js 已安装的前提下，安装器会尽力自动补装缺失的 `claude` / `codex`

## 2. 运行时提示无法自动安装目标 CLI

例如：

```text
Cannot install claude CLI automatically because npm is not installed.
```

或：

```text
Node.js version is too old to install codex CLI.
```

这表示：

- `ccc` 已经安装成功
- 但当前机器还不满足所选目标 CLI 的运行依赖

处理方式：

```powershell
npm install -g @anthropic-ai/claude-code
npm install -g @openai/codex
```

如果仍然失败，优先检查：

```powershell
node --version
npm --version
Get-Command claude -ErrorAction SilentlyContinue
Get-Command codex -ErrorAction SilentlyContinue
```

## 3. 配置文件无法解析

配置文件路径：

```text
$env:USERPROFILE\.ccc\config.json
```

### 快速检查

```powershell
Get-Content $env:USERPROFILE\.ccc\config.json -Raw | ConvertFrom-Json
```

如果报错，通常是：

- JSON 语法错误
- 文件编码异常
- 手动编辑时多了逗号或缺了引号

### 推荐修复方式

先备份旧文件：

```powershell
Copy-Item $env:USERPROFILE\.ccc\config.json $env:USERPROFILE\.ccc\config.json.backup -ErrorAction SilentlyContinue
```

然后直接用：

```powershell
ccc --edit
```

或重新创建：

```powershell
Set-Content -Path $env:USERPROFILE\.ccc\config.json -Value "[]" -Encoding utf8
ccc --add
```

### 正确的最小配置示例

```json
[
  {
    "name": "ZHIPU AI",
    "env": {
      "ANTHROPIC_BASE_URL": "https://open.bigmodel.cn/api/anthropic",
      "ANTHROPIC_AUTH_TOKEN": "your-api-key",
      "ANTHROPIC_MODEL": "glm-5",
      "ANTHROPIC_SMALL_FAST_MODEL": "glm-4.7"
    }
  }
]
```

不要把真实 API Key 写进文档、截图或 issue。

## 4. 如何验证配置是否有效

```powershell
ccc --list
ccc --show
ccc --validate
```

说明：

- `--show` 只会部分显示密钥
- `--validate` 会按当前 `command` 校验 `ANTHROPIC_*` 或 `OPENAI_*`

## 5. `ccc` 命令无法使用

先检查 PowerShell profile 和 PATH 是否生效：

```powershell
Get-Command ccc -ErrorAction SilentlyContinue
$env:PATH -split ';'
```

如果安装刚完成但当前窗口还没加载新环境，重开一个 PowerShell 窗口再试。

## 6. 完全重装

```powershell
ccc --uninstall
irm https://raw.githubusercontent.com/LiukerSun/cc-cli/main/install.ps1 | iex
```

如果只想重建配置文件：

```powershell
Remove-Item $env:USERPROFILE\.ccc\config.json -Force
ccc --add
```
