import subprocess
import logging
import sys
from mcp.server.fastmcp import FastMCP
from typing import Tuple
import os
import platform

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[
        logging.StreamHandler(sys.stdout),
        logging.FileHandler('mcp_bash_server.log')
    ]
)
logger = logging.getLogger(__name__)
# Create a Model-Context-Protocol (MCP) server
mcp = FastMCP("Shell", host="0.0.0.0", port=8080)
logger.info("MCP Shell server initialized")
logger.info(f"Server will run on {mcp.settings.host}:{mcp.settings.port}")

# Global variable for working directory
GLOBAL_CWD = os.getcwd()  # Default to current directory
logger.info(f"Initial working directory: {GLOBAL_CWD}")

@mcp.tool()
async def get_os_info() -> str:
    """
    Get the operating system of the server.

    Returns:
        The operating system name (e.g., "Linux", "Windows", "Darwin").
    """
    logger.info("Request for OS info received.")
    os_info = platform.system()
    logger.info(f"Returning OS info: {os_info}")
    return os_info

@mcp.tool()
async def set_cwd(path: str) -> str:
    """
    Set the global working directory for shell commands.

    Args:
        path: The absolute path to use as the new working directory.

    Returns:
        A confirmation message.
    """
    global GLOBAL_CWD
    logger.info(f"Attempting to set working directory to: {path}")

    if not os.path.isdir(path):
        error_msg = f"Invalid directory: {path}"
        logger.error(error_msg)
        raise ValueError(error_msg)

    GLOBAL_CWD = path
    logger.info(f"Working directory successfully set to: {GLOBAL_CWD}")
    return f"Working directory set to: {GLOBAL_CWD}"

@mcp.tool()
async def execute_shell(cmd: str) -> Tuple[str, str]:
    """
    Run a shell command in the global working directory.

    Args:
        cmd: The shell command to execute.

    Returns:
        A tuple (stdout, stderr) from the command execution.
    """
    logger.info(f"Executing shell command: {cmd}")
    logger.info(f"Working directory: {GLOBAL_CWD}")

    process = subprocess.Popen(
        cmd,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
        shell=True,
        cwd=GLOBAL_CWD
    )
    stdout, stderr = process.communicate()

    logger.info(f"Command completed with return code: {process.returncode}")
    if stdout:
        logger.info(f"stdout: {stdout[:200]}{'...' if len(stdout) > 200 else ''}")
    if stderr:
        logger.warning(f"stderr: {stderr[:200]}{'...' if len(stderr) > 200 else ''}")

    return stdout, stderr

if __name__ == "__main__":
    logger.info("Starting MCP Shell server...")

    # 使用 streamable-http 传输方式，这样可以通过HTTP访问
    try:
        logger.info("Starting server with streamable-http transport...")
        mcp.run(transport="streamable-http")
    except KeyboardInterrupt:
        logger.info("Server stopped by user")
    except Exception as e:
        logger.error(f"Server error: {e}")
        raise