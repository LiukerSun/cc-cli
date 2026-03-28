# CC-CLI 项目结构

当前仓库处于 Go 重构阶段，目录分成三部分：

1. 新的 Go 实现
2. 过渡期安装与文档
3. 过渡期兼容入口

```text
cc-cli/
├── .github/
│   └── workflows/
│       ├── ci.yml                # Linux / Windows CI
│       └── release.yml           # tag 发布到 GitHub Releases
├── cmd/
│   └── ccc/                      # Go CLI 入口
├── internal/
│   ├── app/                      # 命令分发
│   ├── buildinfo/                # 版本注入
│   ├── config/                   # 新配置 schema 与迁移读取
│   ├── deps/                     # Node/npm/claude/codex 检查与自动安装
│   ├── legacy/                   # 旧路径探测
│   ├── platform/                 # 目录布局与平台路径
│   ├── runner/                   # 运行计划与命令执行
│   └── sync/                     # Claude / Codex 外部配置同步
├── docs/
│   ├── configuration.md          # 当前配置模型
│   ├── go-refactor.md            # 重构里程碑说明
│   ├── USAGE.md                  # 当前 Go CLI 使用方式
│   └── windows-troubleshooting.md
├── tests/
│   ├── install_best_effort.sh    # 安装 / 卸载 smoke test
│   ├── check_version.sh          # VERSION 与 tag 一致性校验
│   ├── install_requires_node.sh  # 安装器不再依赖 Node.js 的验证
│   ├── install_windows.ps1       # Windows 安装器 smoke test
│   └── test.sh                   # 当前 Go CLI smoke test
├── bin/
│   ├── ccc                       # 兼容 wrapper，转发到已安装 Go 二进制
│   └── ccc.ps1                   # Windows 兼容 wrapper
├── .goreleaser.yaml              # Go release 打包配置
├── config.example.json           # 新配置示例
├── install.ps1                   # Windows 薄安装器
├── install.sh                    # Unix 薄安装器
├── README.md                     # 当前主文档
└── VERSION                       # 当前版本号
```

## 关键文件说明

- `cmd/ccc/main.go`
  - Go CLI 可执行入口
- `internal/app`
  - 命令路由和用户可见命令行为
- `internal/config`
  - 新配置 schema、旧配置导入、持久化
- `internal/deps`
  - `node` / `npm` / `claude` / `codex` 检查与按需自动安装
- `internal/runner`
  - 构造运行计划、透传参数、实际执行目标 CLI
- `internal/sync`
  - 同步 `~/.claude/settings.json` 和 `~/.codex/*`
- `install.sh` / `install.ps1`
  - 薄安装器
  - 本地仓库里优先直接构建
  - release 场景下载预编译二进制
- `.github/workflows/ci.yml`
  - Linux / Windows 基础 CI
- `.github/workflows/release.yml`
  - tag 推送后执行 GoReleaser 发布
- `bin/ccc` / `bin/ccc.ps1`
  - 兼容 wrapper
  - 不再承载主业务逻辑

## 当前主线

如果修改的是以下内容，通常应该优先在 Go 实现中完成，而不是继续扩展旧脚本：

- 配置模型
- 运行行为
- 依赖检测
- CLI 自动安装
- 外部配置同步
- 安装路径与发布流程

## 文档同步约定

如果修改以下行为，通常需要同步更新 README 或 `docs/`：

- 命令行参数
- 安装行为
- 默认目录布局
- 配置格式
- 外部配置同步行为
- 运行时依赖要求
