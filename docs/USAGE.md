# 使用指南

## 查看帮助

```bash
cc --help
```

## 查看所有模型

```bash
cc --list
```

## 交互式选择

```bash
cc
```

使用上下键选择，按 Enter 确认。

## 查看当前模型

```bash
cc --current
```

## 查看 API Keys

```bash
cc --show-keys
```

## 编辑配置

```bash
cc --edit
```

## 添加新模型

```bash
cc --add
```

## 使用特定模型启动

```bash
cc 2
```

## Bypass 模式启动

```bash
cc -y 2
```

## 传递参数给 Claude

```bash
cc 1 -- --help
cc 2 -- "写一个 hello world 程序"
```

## 仅设置环境变量

```bash
cc -e 2
source /tmp/cc-model-env.sh
claude
```
