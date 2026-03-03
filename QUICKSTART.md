# 快速开始

## 一键安装

```bash
curl -fsSL https://raw.githubusercontent.com/LiukerSun/cc-cli/main/install.sh | bash
```

## 验证安装

```bash
cc --version
cc --help
```

## 基本使用

```bash
# 交互式选择
cc

# 列出所有模型
cc --list

# 查看当前模型
cc --current

# 编辑配置
cc --edit

# 添加新模型
cc --add
```

## 下一步

1. 编辑配置文件添加你的 API keys
```bash
cc --edit
```

2. 重新加载 shell
```bash
source ~/.zshrc  # 或 ~/.bashrc
```

3. 开始使用
```bash
cc
```

## 文档

- [完整文档](https://github.com/LiukerSun/cc-cli#readme)
- [配置指南](https://github.com/LiukerSun/cc-cli/blob/main/docs/configuration.md)
- [使用指南](https://github.com/LiukerSun/cc-cli/blob/main/docs/usage.md)

## 问题反馈

https://github.com/LiukerSun/cc-cli/issues
