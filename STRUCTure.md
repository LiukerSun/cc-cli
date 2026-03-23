# CC-CLI 项目结构

```text
cc-cli/
├── bin/
│   ├── ccc                        # Bash 主脚本
│   └── ccc.ps1                    # PowerShell 主脚本
├── docs/
│   ├── configuration.md           # 配置说明
│   ├── USAGE.md                   # 使用指南
│   └── windows-troubleshooting.md # Windows 故障排除
├── tests/
│   ├── install_best_effort.sh     # 安装器 best-effort 行为测试
│   ├── install_requires_node.sh   # 安装器必须先有 Node.js 的测试
│   └── test.sh                    # 旧测试脚本
├── CHANGELOG.md                   # 更新日志
├── CONTRIBUTING.md                # 贡献指南
├── LICENSE                        # MIT 许可证
├── README.md                      # 主文档
├── STRUCTure.md                   # 本文件
├── VERSION                        # 当前版本号
├── config.example.json            # 示例配置
├── install.ps1                    # PowerShell 安装脚本
├── install.sh                     # Bash 安装脚本
└── Makefile                       # 辅助构建文件
```

## 关键文件说明

- `bin/ccc`
  - Linux / macOS 下的主入口
  - 负责模型选择、配置读写、Claude/Codex 启动与同步
- `bin/ccc.ps1`
  - Windows PowerShell 下的主入口
  - 与 Bash 版本保持尽量一致的行为
- `install.sh`
  - Bash 安装器
  - 负责复制脚本、写入 PATH、检查 Node.js、尽力补装缺失 CLI
- `install.ps1`
  - PowerShell 安装器
  - 负责安装脚本、配置 PowerShell 包装器和 PATH
- `docs/`
  - 存放面向用户的详细文档
- `tests/`
  - 存放安装流程与脚本行为的验证脚本

## 文档同步约定

如果修改以下行为，通常需要同步更新 README 或 `docs/`：

- 命令行参数
- 安装前置条件
- 自动安装 / 升级 / 卸载行为
- 配置格式
- Windows 专项问题排查说明
