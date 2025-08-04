# pip install google-genai mcp
# requires Python 3.13+
import os
import asyncio
from google import genai
from google.genai import types
from mcp import ClientSession
from mcp.client.streamable_http import streamablehttp_client
from dotenv import load_dotenv, find_dotenv
load_dotenv(find_dotenv())

# Create Gemini instance LLM class
client = genai.Client(
    api_key=os.getenv("GEMINI_API_KEY"),
    http_options=types.HttpOptions(base_url=os.getenv("GEMINI_BASE_URL")),
)

remote_url = "https://mcp.deepwiki.com/mcp"


async def run():
    async with streamablehttp_client(remote_url) as (read, write, _):
        async with ClientSession(read, write) as session:
            await session.initialize()
            # Initialize conversation history using simple tuples
            config = genai.types.GenerateContentConfig(
                temperature=0,
                tools=[session],
            )
            print("Agent is ready. Type 'exit' to quit.")
            chat = client.aio.chats.create(
                model="gemini-2.5-flash", config=config
            )
            while True:
                user_input = input("You: ")
                if user_input.lower() == "exit":
                    print("Exiting chat.")
                    break

                # Append user message to history
                response = await chat.send_message(user_input)
                if len(response.automatic_function_calling_history) > 0:
                    if (
                        response.automatic_function_calling_history[0].parts[0].text
                        == user_input
                    ):
                        response.automatic_function_calling_history.pop(0)
                    for call in response.automatic_function_calling_history:
                        if call.parts[0].function_call:
                            print(f"MCP call: {call.parts[0].function_call}")
                        elif call.parts[0].function_response:
                            print(
                                f"MCP response: {call.parts[0].function_response.response["result"].content[0].text}"
                            )
                print(f"Assistant: {response.text}")


if __name__ == "__main__":
    asyncio.run(run())
