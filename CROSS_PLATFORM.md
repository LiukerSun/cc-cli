# 跨平台支持说明

## 支持的平台

CC-CLI 现在支持三个主要平台：

- **macOS** - 使用 Bash 脚本
- **Linux** - 使用 Bash 脚本
- **Windows** - 使用 PowerShell 脚本

## 平台差异

### 安装

**macOS / Linux:**
```bash
curl -fsSL https://raw.githubusercontent.com/LiukerSun/cc-cli/main/install.sh | bash
```

**Windows PowerShell:**
```powershell
# 方式 1: 直接执行（需要管理员权限）
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
irm https://raw.githubusercontent.com/LiukerSun/cc-cli/main/install.ps1 | iex

# 方式 2: 下载后执行
Invoke-WebRequest -Uri https://raw.githubusercontent.com/LiukerSun/cc-cli/main/install.ps1 -OutFile install.ps1
.\install.ps1
```

### 配置文件位置

- **macOS / Linux**: `~/.cc-config.json`
- **Windows**: `%USERPROFILE%\.cc-config.json`

### 环境变量文件

- **macOS / Linux**: `/tmp/cc-model-env.sh`
- **Windows**: `%TEMP%\cc-model-env.ps1`

### 命令差异

**Bash (macOS/Linux):**
- 长选项使用双破折号: `cc --list`
- 短选项使用单破折号: `cc -l`

**PowerShell (Windows):**
- 长选项支持双破折号: `cc --list`
- 短选项支持单破折号: `cc -l`
- 支持别名: `cc -list`, `cc -edit`

### 编辑器

**macOS / Linux:**
- 默认: vim
- 自定义: `export EDITOR=nano`

**Windows:**
- 默认: notepad
- 自定义: `$env:EDITOR = "code"`

## 功能对比

| 功能 | macOS/Linux | Windows |
|------|------------|---------|
| 交互式选择 | ✅ | ✅ |
| 列出模型 | ✅ | ✅ |
| 查看当前 | ✅ | ✅ |
| 编辑配置 | ✅ | ✅ |
| 添加模型 | ✅ | ✅ |
| 查看密钥 | ✅ | ✅ |
| Bypass 模式 | ✅ | ✅ |
| 彩色输出 | ✅ | ✅ |
| API Key 隐藏 | ✅ | ✅ |

## 测试平台

- ✅ macOS (Bash 3.2+)
- ✅ Linux (Bash 4.0+)
- ✅ Windows 10/11 (PowerShell 5.1+)

## 注意事项

### Windows 用户

1. **执行策略**: PowerShell 默认禁止运行脚本，需要先设置执行策略
   ```powershell
   Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
   ```

2. **管理员权限**: 某些安装步骤可能需要管理员权限

3. **Claude CLI**: 确保 Claude CLI 在 PATH 中可用

### macOS / Linux 用户

1. **Bash 版本**: 确保使用 Bash 3.2 或更高版本

2. **文件权限**: 脚本需要可执行权限
   ```bash
   chmod +x ~/bin/cc
   ```

## 故障排除

### Windows: "无法加载文件，因为在此系统上禁止运行脚本"

```powershell
# 解决方法：设置执行策略
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### Windows: "cc 命令未找到"

```powershell
# 解决方法：重启 PowerShell 或运行
$env:PATH += ";$env:USERPROFILE\bin"
```

### macOS/Linux: "权限被拒绝"

```bash
# 解决方法：添加执行权限
chmod +x ~/bin/cc
```

### macOS/Linux: "命令未找到"

```bash
# 解决方法：重新加载 shell
source ~/.zshrc  # 或 ~/.bashrc
```

## 安装验证

### macOS / Linux
```bash
# 检查安装
which cc
cc --version
cc --list
```

### Windows
```powershell
# 检查安装
Get-Command cc
cc -Version
cc -List
```

## 获取帮助

- **文档**: https://github.com/LiukerSun/cc-cli
- **Issues**: https://github.com/LiukerSun/cc-cli/issues
- **讨论**: https://github.com/LiukerSun/cc-cli/discussions
