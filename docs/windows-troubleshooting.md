# Windows 故障排除

本页覆盖当前 Go 版 `ccc` 在 Windows / PowerShell 环境里最常见的问题：

1. 安装器无法完成安装
2. 运行时无法补装 `claude` / `codex`
3. 配置文件解析或迁移失败
4. `ccc` 或当前 profile 命令无法使用

## 1. 安装器无法完成安装

当前 `install.ps1` 只负责安装 `ccc.exe`。它不再要求机器先安装 Node.js。

标准安装方式：

```powershell
irm https://raw.githubusercontent.com/LiukerSun/cc-cli/main/install.ps1 | iex
```

安装完成后先确认：

```powershell
ccc version
ccc help
```

### 常见原因

- PowerShell 执行策略或会话环境异常
- 网络无法访问 GitHub Releases
- 架构或系统环境不受支持
- `ccc.exe` 已安装，但安装目录还没进入当前窗口的 `PATH`

### 快速检查

```powershell
Get-Command ccc -ErrorAction SilentlyContinue
$env:PATH -split ';'
Test-Path "$env:LOCALAPPDATA\Programs\ccc\bin\ccc.exe"
```

当前默认安装目录：

```text
%LOCALAPPDATA%\Programs\ccc\bin\ccc.exe
```

如果文件已经存在但命令不可用，先开一个新的 PowerShell 窗口再试。

## 2. 运行时无法自动安装目标 CLI

例如：

```text
dependency check failed: cannot install claude CLI automatically because npm is not installed
```

或：

```text
dependency check failed: node version 16.0.0 is too old to install claude CLI
```

这表示：

- `ccc` 本身已经安装成功
- 但运行当前 profile 需要的 `claude` 或 `codex` 还不存在
- `ccc run` 尝试自动补装时，发现 Node.js / npm 环境不满足要求

### 版本要求

- Claude CLI: Node.js `>= 18.0.0`
- Codex CLI: Node.js `>= 16.0.0`

### 快速检查

```powershell
node --version
npm --version
Get-Command claude -ErrorAction SilentlyContinue
Get-Command codex -ErrorAction SilentlyContinue
ccc doctor
```

### 手动安装

```powershell
npm install -g @anthropic-ai/claude-code
npm install -g @openai/codex
```

如果你不想让 `ccc` 写入外部工具配置，可以在创建 profile 时加 `--no-sync`。

## 3. 配置文件解析或迁移失败

当前 Windows 默认配置路径：

```text
%APPDATA%\ccc\config.json
```

旧版本配置如果存在，`ccc` 仍会尝试读取：

```text
%USERPROFILE%\.ccc\config.json
%USERPROFILE%\.cc-config.json
```

### 快速检查当前读取来源

```powershell
ccc config path
ccc config show
```

如果需要结构化查看：

```powershell
ccc config show | ConvertFrom-Json
```

### 推荐迁移方式

```powershell
ccc config migrate
```

这会把当前读取到的旧配置写入新路径。

### 如果配置文件损坏

先备份：

```powershell
Copy-Item "$env:APPDATA\ccc\config.json" "$env:APPDATA\ccc\config.json.backup" -ErrorAction SilentlyContinue
```

然后优先用命令重新创建 profile，而不是手动拼 JSON：

```powershell
ccc profile add `
  --api-key your-api-key `
  --preset anthropic `
  --name "Claude Official"
```

### 当前最小配置示例

```json
{
  "version": 1,
  "current_profile": "claude-official",
  "profiles": [
    {
      "id": "claude-official",
      "name": "Claude Official",
      "provider": "anthropic",
      "command": "claude",
      "base_url": "https://api.anthropic.com",
      "api_key": "your-api-key",
      "model": "claude-3-7-sonnet",
      "sync_external": true
    }
  ]
}
```

不要把真实 API Key 写进文档、截图或 issue。

## 4. 如何验证当前 profile 是否正确

```powershell
ccc profile list
ccc current
ccc run --dry-run
ccc sync --dry-run
```

建议关注：

- 当前读取的是哪个配置文件
- `current_profile` 是否存在
- `command` 是 `claude` 还是 `codex`
- `base_url` / `model` 是否符合你的实际 provider
- 外部同步目标是否符合预期

## 5. `ccc` 或当前 profile 命令无法使用

先区分两类问题：

1. `ccc` 自己找不到
2. `ccc` 找得到，但当前 profile 的 `claude` / `codex` 找不到

### 检查 `ccc`

```powershell
Get-Command ccc -ErrorAction SilentlyContinue
Test-Path "$env:LOCALAPPDATA\Programs\ccc\bin\ccc.exe"
```

### 检查当前 profile 目标命令

```powershell
ccc doctor
ccc current
Get-Command claude -ErrorAction SilentlyContinue
Get-Command codex -ErrorAction SilentlyContinue
```

如果 `ccc doctor` 提示当前 profile 命令不在 `PATH`，优先修正 Node/npm 和目标 CLI 安装状态。

## 6. 完全重装

如果你要完整重装二进制，建议先把安装脚本保存到本地，再执行卸载和安装：

```powershell
Invoke-WebRequest https://raw.githubusercontent.com/LiukerSun/cc-cli/main/install.ps1 -OutFile .\install.ps1
powershell -ExecutionPolicy Bypass -File .\install.ps1 -Action uninstall
powershell -ExecutionPolicy Bypass -File .\install.ps1
```

如果你只想重建配置文件：

```powershell
Remove-Item "$env:APPDATA\ccc\config.json" -Force -ErrorAction SilentlyContinue
ccc profile add `
  --api-key your-api-key `
  --preset anthropic `
  --name "Claude Official"
```
