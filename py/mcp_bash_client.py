import asyncio
import logging
from fastmcp import Client

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

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

async def test_bash_server():
    logger.info("连接到MCP Bash服务器...")
    async with mcp_client:
        logger.info("连接成功！")
        tools_result = await mcp_client.session.list_tools()
        logger.info(f"可用工具: {[tool.name for tool in tools_result.tools]}")
        logger.info("测试设置工作目录...")
        result = await mcp_client.session.call_tool("set_cwd", {"path": "/tmp"})
        logger.info(f"设置工作目录结果: {result}")
        logger.info("测试执行bash命令...")
        result = await mcp_client.session.call_tool("execute_bash", {"cmd": "pwd"})
        logger.info(f"执行pwd命令结果: {result}")
        logger.info("测试执行ls命令...")
        result = await mcp_client.session.call_tool("execute_bash", {"cmd": "ls -la"})
        logger.info(f"执行ls命令结果: {result}")

async def interactive_mode():
    logger.info("启动交互模式...")
    async with mcp_client:
        logger.info("连接成功！输入 'exit' 退出")
        while True:
            try:
                command = input("bash> ")
                if command.lower() == 'exit':
                    break
                if command.strip():
                    logger.info(f"执行命令: {command}")
                    result = await mcp_client.session.call_tool("execute_bash", {"cmd": command})
                    stdout, stderr = result.content
                    if stdout and hasattr(stdout, 'strip') and stdout.strip():
                        print(f"输出: {stdout.strip()}")
                    elif stdout and hasattr(stdout, 'text'):
                        print(f"输出: {stdout.text}")
                    if stderr and hasattr(stderr, 'strip') and stderr.strip():
                        print(f"错误: {stderr.strip()}")
                    elif stderr and hasattr(stderr, 'text'):
                        print(f"错误: {stderr.text}")
            except KeyboardInterrupt:
                break
            except Exception as e:
                logger.error(f"执行命令时出错: {e}")

if __name__ == "__main__":
    import sys
    if len(sys.argv) > 1 and sys.argv[1] == "test":
        asyncio.run(test_bash_server())
    else:
        asyncio.run(interactive_mode())
