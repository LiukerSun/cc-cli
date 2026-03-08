# 贡献指南

感谢你考虑为 CC-CLI 做出贡献！

## 开发环境设置

1. Fork 本仓库
```bash
git clone https://github.com/LiukerSun/cc-cli.git
cd cc-cli
```

2. 安装依赖
```bash
# 运行安装脚本
./install.sh
```

3. 创建功能分支
```bash
git checkout -b feature
```

4. 进行更改
```bash
git add .
git commit -m "Your changes"
```

## 代码风格

- Shell 脚本遵循 [Google Shell Style Guide](https://google.github.io/styleguide/shell.xml)
- 使用有意义的变量名和函数名
- 添加注释解释复杂逻辑
- 保持函数简短和专注

## 提交 Bug 报告

请在提交 Issue 之前，请搜索现有的 issues。

## 功能请求

我们非常欢迎新功能！请先开一个 issue 讨论。

## 提交 Pull Request

1. Fork 本仓库
2. 创建你的功能分支
```bash
git checkout -b feature
```

3. 进行更改并编写测试
```bash
git add .
git commit -m "Add your feature"
```

4. 推送到你的分支并创建一个 Pull Request
```bash
git push origin feature-branch
```

## 代码审查

所有提交都需要通过代码审查。我们使用 GitHub Actions 进行 CI 检查。

## 社区准则

- 尊重所有贡献者
- 欢迎建设性讨论
- 保持友好和专业
