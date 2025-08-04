# pip install google-genai fastmcp
# requires Python 3.13+
import os
import asyncio
from datetime import datetime
from google import genai
from google.genai import types
from fastmcp import Client
from dotenv import load_dotenv, find_dotenv
load_dotenv(find_dotenv())

# Create Gemini instance LLM class
client = genai.Client(
    api_key=os.getenv("GEMINI_API_KEY"),
    http_options=types.HttpOptions(base_url=os.getenv("GEMINI_BASE_URL")),
)

mcp_client = Client(
    {
        "mcpServers": {
            "shell": {
                "transport": "streamable-http",
                "url": "http://localhost:8080/mcp",
            }
        }
    }
)

async def run():
    async with mcp_client:
        try:
            # Correctly call the tool using the session
            tool_response = await mcp_client.session.call_tool("get_os_info", {})
            os_info = tool_response.content
            print(f"🤖 Shell server is running on: {os_info}")
        except Exception as e:
            print(f"Error connecting to shell server: {e}")
            print("Please ensure the server is running and accessible.")
            return

        config = genai.types.GenerateContentConfig(
            temperature=0,
            tools=[mcp_client.session],
            system_instruction=f"""Very important: The user's timezone is {datetime.now().strftime("%Z")}. The current date is {datetime.now().strftime("%Y-%m-%d")}.
You are connected to a shell environment. The operating system is {os_info}.
You must use shell commands that are compatible with {os_info}. For example, if the OS is Windows, use 'dir' instead of 'ls'.
Any dates before this are in the past, and any dates after this are in the future. When dealing with modern entities/companies/people, and the user asks for the 'latest', 'most recent', 'today's', etc. don't assume your knowledge is up to date;
You can and should speak any language the user asks you to speak or use the language of the user.""",
        )
        print("🤖 Gemini MCP Agent Ready")
        print("Type 'exit' to quit\n")
        chat = client.aio.chats.create(model="gemini-2.5-flash", config=config)
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
    asyncio.run(run())