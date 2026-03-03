# 安装指南

## 系统要求

- **操作系统**: macOS, Linux, Windows (WSL/Git Bash)
- **Shell**: Bash 3.2 或更高版本
- **可选**: Claude CLI (用于启动 Claude)

## 安装方式

### 方式 1: 一键安装（推荐）

```bash
curl -fsSL https://raw.githubusercontent.com/LiukerSun/cc-cli/main/install.sh | bash
```

### 方式 2: 手动安装

```bash
# 克隆仓库
git clone https://github.com/LiukerSun/cc-cli.git
cd cc-cli

# 运行安装脚本
chmod +x install.sh
./install.sh
```

### 方式 3: 直接下载

```bash
# 创建目录
mkdir -p ~/bin

# 下载脚本
curl -fsSL https://raw.githubusercontent.com/LiukerSun/cc-cli/main/bin/cc -o ~/bin/cc

# 添加执行权限
chmod +x ~/bin/cc

# 添加到 PATH
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

## 验证安装

```bash
# 检查命令
which cc

# 查看帮助
cc --help

# 列出模型
cc --list
```

## 配置

### 首次使用

1. **编辑配置文件**
```bash
cc --edit
```

2. **添加 API keys**
在打开的编辑器中，替换 `your-api-key-here` 为你的实际 API key

3. **保存并退出**
- vim: 按 `ESC`，输入 `:wq`，按 `Enter`
- nano: 按 `Ctrl+X`，按 `Y`，按 `Enter`

4. **重新加载 shell**
```bash
source ~/.zshrc  # 或 ~/.bashrc
```

### 添加新模型

```bash
cc --add
```

按提示输入：
- 模型名称
- API Base URL
- API Key
- 主模型名称
- 快速模型名称（可选）

## 故障排除

### 命令未找到

```bash
# 检查 PATH
echo $PATH | grep -o "$HOME/bin"

# 如果没有输出，添加到 PATH
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

### 权限被拒绝

```bash
chmod +x ~/bin/cc
```

### 配置文件损坏

```bash
# 备份并重新创建
cp ~/.cc-config.json ~/.cc-config.json.backup
cc --add
```

### Bash 版本过低

```bash
# 检查 Bash 版本
bash --version

# 如果版本 < 3.2，升级 Bash
# macOS: brew install bash
# Linux: sudo apt-get install bash
```

## 卸载

```bash
./install.sh --uninstall

# 或手动卸载
rm -f ~/bin/cc
rm -rf ~/.cc-cli
# 配置文件会保留，可手动删除
rm -f ~/.cc-config.json
```

## 更新

```bash
# 重新运行安装脚本
curl -fsSL https://raw.githubusercontent.com/LiukerSun/cc-cli/main/install.sh | bash

# 或使用 git
cd cc-cli
git pull
./install.sh
```

## 下一步

安装完成后：

1. 添加你的 API keys: `cc --edit`
2. 查看所有模型: `cc --list`
3. 开始使用: `cc`

## 获取帮助

- **文档**: https://github.com/LiukerSun/cc-cli#readme
- **Issues**: https://github.com/LiukerSun/cc-cli/issues
- **示例**: https://github.com/LiukerSun/cc-cli/tree/main/docs
