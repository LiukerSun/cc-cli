# 使用指南

## 前置条件

- `install.sh` / `install.ps1` 运行前要求系统已安装 Node.js
- 当前最低要求：
  - Claude CLI: Node.js `>= 18.0.0`
  - Codex CLI: Node.js `>= 16.0.0`
- `ccc` 在运行目标命令时会按需尝试通过 npm 自动安装缺失的 `claude` / `codex`

## 查看帮助和版本

```bash
ccc --help
ccc --version
```

## 查看和选择模型

```bash
ccc --list
ccc --current
ccc
ccc 2
```

说明：
- `ccc` 会打开交互式菜单
- `ccc 2` 会直接使用模型 `#2`
- 通过 `--` 之后的参数会原样传给目标 CLI

## 常用配置命令

```bash
ccc --add
ccc --edit
ccc --show
ccc --validate
ccc --delete 2
```

说明：
- `--show` 显示已保存的 API Key，输出会做部分隐藏
- `--validate` 会校验配置；在支持 `jq` 的环境下会自动移除无效项

## 传递参数给目标 CLI

```bash
ccc 1 -- --help
ccc 2 -- "Write a hello world program"
ccc 3 -- --version --verbose
```

## Bypass 模式

```bash
ccc --bypass 2
ccc -y 2
```

说明：
- 当所选命令为 `claude` 时，`-y` 会映射为 `--dangerously-skip-permissions`
- 当所选命令为 `codex` 时，`-y` 会映射为 `--dangerously-bypass-approvals-and-sandbox`

## 仅导出环境变量

```bash
ccc --env 2
source ~/.ccc/tmp/cc-model-env.sh
```

如果模型配置的 `command` 是 `claude`，之后通常手动运行：

```bash
claude
```

如果模型配置的 `command` 是 `codex`，则运行：

```bash
codex
```

## 升级和卸载

```bash
ccc --upgrade
ccc --uninstall
```

## 常见运行时提示

### 1. 缺少 Node.js

如果在安装阶段看到：

```text
Node.js is required to install ccc
```

表示当前机器还没有安装 Node.js，安装器会直接退出。
在 macOS / Linux 上，安装器会额外建议优先使用 `nvm` 安装 Node.js，并直接输出安装命令。

### 2. 运行时无法自动安装目标 CLI

如果在启动模型时看到：

```text
Cannot install claude CLI automatically because Node.js is not installed.
```

或：

```text
Cannot install codex CLI automatically because npm is not installed.
```

表示 `ccc` 本身已经安装成功，但当前机器还不满足目标 CLI 的运行依赖；先安装或升级 Node.js / npm，再重新运行 `ccc`。
