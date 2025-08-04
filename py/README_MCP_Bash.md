# MCP Bash 服务器使用指南

## 概述

这是一个基于 Model Context Protocol (MCP) 的 Bash 命令执行服务器，允许通过 HTTP 接口远程执行 bash 命令。

## 功能特性

- 🔧 **远程 Bash 命令执行**: 通过 HTTP 接口执行 bash 命令
- 📁 **工作目录管理**: 支持设置和切换工作目录
- 📝 **详细日志记录**: 完整的运行时日志，包括命令执行和错误信息
- 🌐 **HTTP 传输**: 使用 streamable-http 传输协议
- 🔒 **安全执行**: 在指定工作目录中执行命令

## 服务器配置

### 端口和主机设置

服务器默认运行在 `0.0.0.0:8080`，可以通过修改 `mcp_bash_server.py` 中的配置来更改：

```python
mcp = FastMCP("Bash", host="0.0.0.0", port=8080)
```

### 可用的传输方式

FastMCP 支持以下传输方式：
- `stdio` (默认) - 标准输入输出
- `sse` - Server-Sent Events
- `streamable-http` - HTTP 流式传输

## 使用方法

### 1. 启动服务器

```bash
uv run mcp_bash_server.py
```

服务器启动后会显示：
```
2025-08-04 11:12:22,787 - __main__ - INFO - MCP Bash server initialized
2025-08-04 11:12:22,787 - __main__ - INFO - Server will run on 0.0.0.0:8080
INFO:     Uvicorn running on http://0.0.0.0:8080 (Press CTRL+C to quit)
```

### 2. 使用客户端

#### 测试模式
```bash
uv run mcp_bash_client.py test
```

#### 交互模式
```bash
uv run mcp_bash_client.py
```

在交互模式中，你可以输入 bash 命令：
```
bash> pwd
输出: /private/tmp

bash> ls -la
输出: total 0
drwxrwxrwt  5 root  wheel  160 Aug  4 09:42 .
drwxr-xr-x  6 root  wheel  192 Jul 28 21:22 ..
...
```

## API 接口

### 可用工具

1. **set_cwd(path: str) -> str**
   - 设置工作目录
   - 参数: `path` - 绝对路径
   - 返回: 确认消息

2. **execute_bash(cmd: str) -> Tuple[str, str]**
   - 执行 bash 命令
   - 参数: `cmd` - 要执行的 shell 命令
   - 返回: (stdout, stderr) 元组

### HTTP 端点

- **端点**: `http://localhost:8080/mcp`
- **协议**: streamable-http
- **方法**: POST/GET/DELETE

## 客户端配置

客户端使用以下配置连接到服务器：

```python
mcp_client = Client(
    {
        "mcpServers": {
            "bash": {
                "transport": "streamable-http",
                "url": "http://localhost:8080/mcp",
            }
        }
    }
)
```

## 日志记录

服务器提供详细的日志记录：

- **控制台输出**: 实时显示服务器状态和命令执行情况
- **文件日志**: 保存到 `mcp_bash_server.log` 文件
- **日志级别**: INFO 级别，包含时间戳和详细信息

### 日志示例

```
2025-08-04 11:12:22,787 - __main__ - INFO - MCP Bash server initialized
2025-08-04 11:12:22,787 - __main__ - INFO - Server will run on 0.0.0.0:8080
2025-08-04 11:12:22,787 - __main__ - INFO - Initial working directory: /Users/zliu/git/mcp_misc/py
2025-08-04 11:12:22,788 - __main__ - INFO - Starting MCP Bash server...
2025-08-04 11:12:22,788 - __main__ - INFO - Starting server with streamable-http transport...
```

## 安全注意事项

1. **工作目录限制**: 命令在指定的工作目录中执行
2. **错误处理**: 无效目录会抛出异常
3. **日志记录**: 所有命令执行都有详细日志
4. **网络访问**: 服务器绑定到 `0.0.0.0`，确保网络可访问

## 故障排除

### 常见问题

1. **连接失败**: 确保服务器正在运行在正确的端口
2. **404 错误**: 确保使用正确的端点 `/mcp`
3. **权限错误**: 检查工作目录的读写权限
4. **命令执行失败**: 查看日志文件了解详细错误信息

### 调试命令

```bash
# 检查服务器状态
curl -v http://localhost:8080/mcp

# 查看日志文件
tail -f mcp_bash_server.log

# 检查端口占用
lsof -i :8080
```

## 扩展功能

可以轻松扩展服务器功能：

1. **添加新工具**: 使用 `@mcp.tool()` 装饰器
2. **添加资源**: 使用 `@mcp.resource()` 装饰器
3. **自定义配置**: 修改 FastMCP 初始化参数
4. **认证机制**: 添加 OAuth 或其他认证方式

## 依赖项

- `fastmcp>=2.11.0`
- `mcp[cli]>=1.12.3`
- Python 3.12+
