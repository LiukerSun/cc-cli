# 贡献指南

感谢你考虑为 CC-CLI 做出贡献。

## 开发环境

建议准备以下工具：

- Git
- Bash
- Node.js
- npm

说明：

- 当前安装器要求机器已安装 Node.js
- 如果你需要验证自动补装逻辑，Node.js / npm 是必需的

## 克隆仓库

```bash
git clone https://github.com/LiukerSun/cc-cli.git
cd cc-cli
```

## 创建分支

```bash
git checkout -b feature/your-change
```

## 本地验证

当前仓库至少应执行这些检查：

```bash
bash -n bin/ccc install.sh
./tests/install_requires_node.sh
./tests/install_best_effort.sh
```

如果你修改了 PowerShell 安装或运行逻辑，建议额外在 Windows / PowerShell 环境中手动验证：

- `install.ps1`
- `bin/ccc.ps1`
- `ccc --help`
- `ccc --add`

## 提交规范

本项目使用约定式提交风格，常见前缀包括：

- `feat`
- `fix`
- `docs`
- `test`
- `chore`

示例：

```bash
git commit -m "fix: require node before installer runs"
```

## 提交 Pull Request

1. 基于最新 `main` 创建分支
2. 完成修改并补充必要测试
3. 如果行为变化影响用户，请同步更新 README 和 `docs/`
4. 推送分支并创建 Pull Request

```bash
git push origin feature/your-change
```

## 代码风格

- Shell 脚本保持 POSIX/Bash 兼容性意识，避免不必要的高版本特性
- PowerShell 改动需兼顾 Windows PowerShell 5.x 的兼容性
- 注释只保留必要说明
- 用户可见提示文案要和实际行为一致

## Bug 报告与功能请求

- Bug: <https://github.com/LiukerSun/cc-cli/issues>
- Feature request: <https://github.com/LiukerSun/cc-cli/issues/new>
