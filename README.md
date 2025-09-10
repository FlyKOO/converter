# 转换器（converter）

一个用于演示如何处理 Yellowstone gRPC 交易数据的 Go 语言示例项目。

本仓库包含连接 Geyser、记录日志以及解析交易详情的工具代码，所有注释均已翻译为中文，方便中文开发者学习和使用。

## 目录结构

- `example/` 示例程序，展示如何订阅并处理交易
- `geyser/` 与 Geyser 节点建立连接的工具
- `logger/` 初始化日志组件
- `shared/` 共享的结构体定义
- `utils/` 交易解析相关的辅助函数

## 使用方法

运行示例程序：

```bash
go run ./example
```

在运行前请将示例中的 `endpointURL` 修改为你自己的 gRPC 服务地址。
