# 贡献指南

感谢你考虑为 CC-CLI 做出贡献。

## 开发环境

建议准备以下工具：

- Git
- Bash
- Go
- Node.js
- npm

说明：

- Go 是当前主实现语言，修改 CLI、配置层、安装路径或同步逻辑时都需要它
- Node.js / npm 不再是安装 `ccc` 本身的前置依赖
- 如果你需要验证 `ccc run` 对 `claude` / `codex` 的自动补装逻辑，Node.js / npm 仍然是必需的

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
make test
make build
```

如果你修改了文档或命令界面，建议额外检查：

```bash
go run ./cmd/ccc help
go run ./cmd/ccc doctor
go run ./cmd/ccc paths --json
```

如果你修改了 PowerShell 安装或 Windows 路径逻辑，建议额外在 Windows / PowerShell 环境中手动验证：

- `powershell -ExecutionPolicy Bypass -File .\install.ps1`
- `ccc version`
- `ccc help`
- `ccc doctor`
- `ccc config path`

仓库 CI 当前会执行：

- Ubuntu: `go test ./...`、`go build ./cmd/ccc`、Unix smoke tests
- Windows: `go test ./...`、`go build ./cmd/ccc`、`tests/install_windows.ps1`

如果你准备发版，额外检查：

```bash
bash tests/check_version.sh
bash tests/check_version.sh v$(cat VERSION)
```

## 提交规范

本项目使用约定式提交风格，常见前缀包括：

- `feat`
- `fix`
- `docs`
- `test`
- `chore`

示例：

```bash
git commit -m "feat: add go-based profile migration command"
```

## 提交 Pull Request

1. 基于最新 `main` 创建分支
2. 完成修改并补充必要测试
3. 如果行为变化影响用户，请同步更新 README 和 `docs/`
4. 推送分支并创建 Pull Request

```bash
git push origin feature/your-change
```

## 发布流程

1. 更新 `VERSION`
2. 确认 `make test` 和 `make build` 通过
3. 合并到 `main`
4. 创建并推送同版本 tag，例如：

```bash
git tag v$(cat VERSION)
git push origin v$(cat VERSION)
```

`release.yml` 会阻止 tag 与 `VERSION` 不一致的发布。

## 代码风格

- Go 代码保持 `gofmt` 输出
- Shell 脚本保持 POSIX/Bash 兼容性意识，避免不必要的高版本特性
- PowerShell 改动需兼顾 Windows PowerShell 5.x 的兼容性
- 注释只保留必要说明
- 用户可见提示文案要和实际行为一致

## 兼容层说明

仓库里仍保留 `bin/ccc` 和 `bin/ccc.ps1` 作为兼容 wrapper，但它们不再承载主业务逻辑。新的行为设计应优先体现在：

- `cmd/ccc`
- `internal/`
- `install.sh`
- `install.ps1`
- `docs/`

## Bug 报告与功能请求

- Bug: <https://github.com/LiukerSun/cc-cli/issues>
- Feature request: <https://github.com/LiukerSun/cc-cli/issues/new>
