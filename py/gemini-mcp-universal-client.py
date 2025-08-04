# pip install google-genai fastmcp python-dotenv
import os
import asyncio
import json
from pathlib import Path
from datetime import datetime
from google import genai
from google.genai import types
from fastmcp import Client
from dotenv import load_dotenv, find_dotenv

# Load environment variables from .env file
load_dotenv(find_dotenv())

# --- Global Capability Registry ---
# This dictionary will hold the information about all discovered tools.
# Format: {'server_name.tool_name': 'Tool description', ...}
CAPABILITY_REGISTRY = {}

async def discover_and_register_capabilities(mcp_client: Client):
    """
    Connects to all servers, gets their tools, and registers them.
    """
    print("🤖 Starting dynamic discovery of all MCP server capabilities...")
    try:
        # list_tools() efficiently queries all servers defined in the client's config.
        all_tools_response = await mcp_client.session.list_tools()
        if not all_tools_response.tools:
            print("⚠️ No tools found on any connected servers.")
            return
        for tool in all_tools_response.tools:
            # The tool name is automatically prefixed with the server name by fastmcp.
            tool_full_name = tool.name
            tool_description = tool.description
            # Register the capability
            CAPABILITY_REGISTRY[tool_full_name] = tool_description
            print(f"  ✅ Discovered and registered tool: {tool_full_name}")
        print("✨ Capability discovery complete!")
    except Exception as e:
        print(f"❌ Error during capability discovery: {e}")
        print("   Please ensure all configured MCP servers are running and accessible.")

def generate_dynamic_system_instruction() -> str:
    """
    Builds the system instruction for the LLM based on the discovered tools.
    """
    if not CAPABILITY_REGISTRY:
        return "You are a helpful assistant. No external tools are available."
    header = f"""You are a powerful AI assistant connected to multiple external systems via tools.
The current date is {datetime.now().strftime("%Y-%m-%d")}.
To use a tool, you must respond with a `ToolCall` object for the corresponding function.
Here are the tools available to you:\n"""
    tool_descriptions = []
    for tool_name, description in CAPABILITY_REGISTRY.items():
        tool_descriptions.append(f"- Function Name: `{tool_name}`\n  Description: {description}")
    return header + "\n\n".join(tool_descriptions)

async def main():
    """
    The main entry point for the universal client.
    """
    print("--- Gemini Universal MCP Client ---")
    # 1. Load Server Configuration
    config_path = Path(__file__).parent / "mcp_servers.json"
    if not config_path.exists():
        print(f"❌ Configuration file not found at: {config_path}")
        print("   Please make sure 'mcp_servers.json' exists.")
        return
    with open(config_path, 'r') as f:
        mcp_config = json.load(f)
    # 2. Initialize Clients
    try:
        gemini_client = genai.Client(
            api_key=os.getenv("GEMINI_API_KEY"),
            http_options=types.HttpOptions(base_url=os.getenv("GEMINI_BASE_URL")),
        )
        mcp_client = Client(mcp_config)
    except Exception as e:
        print(f"❌ Failed to initialize clients: {e}")
        return
    # 3. Discover capabilities and configure the LLM
    async with mcp_client:
        await discover_and_register_capabilities(mcp_client)
        system_instruction = generate_dynamic_system_instruction()
        print("\n--- Generated System Instruction for LLM ---")
        print(system_instruction)
        print("------------------------------------------\n")
        config = genai.types.GenerateContentConfig(
            temperature=0,
            tools=[mcp_client.session],
            system_instruction=system_instruction,
        )
        chat = gemini_client.aio.chats.create(model="gemini-2.5-flash", config=config)
        # 4. Start Interactive Chat
        print("🤖 Universal MCP Agent Ready. Type 'exit' to quit.\n")
        while True:
            user_input = input("You: ")
            if user_input.lower() == "exit":
                print("\nGoodbye!")
                break
            response = await chat.send_message_stream(user_input)
            print("Gemini: ", end="", flush=True)
            async for chunk in response:
                print(chunk.text, end="", flush=True)
            print("\n")

if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        print("\nClient stopped by user.")
