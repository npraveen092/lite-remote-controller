import argparse
import asyncio
import base64
import json
import secrets
import sys
from pathlib import Path

from PIL import Image
from websockets.asyncio.client import connect


async def request(websocket, request_type: str, **payload):
    request_id = secrets.token_hex(8)
    await websocket.send(json.dumps({
        "type": request_type,
        "requestId": request_id,
        **payload,
    }))

    while True:
        raw = await websocket.recv()
        if isinstance(raw, bytes):
            raw = raw.decode("utf-8")
        response = json.loads(raw)
        if response.get("requestId") == request_id or response.get("type") in {"AUTH_OK", "ERROR"}:
            return response


async def run(args):
    uri = f"ws://{args.host}:{args.port}"

    async with connect(uri, max_size=2 * 1024 * 1024) as websocket:
        await websocket.send(json.dumps({"type": "AUTH", "token": args.token}))
        auth = json.loads(await websocket.recv())
        if auth.get("type") != "AUTH_OK":
            raise RuntimeError(f"Authentication failed: {auth}")

        print("Connected.")
        print(json.dumps(auth.get("system", {}), indent=2))
        print("\nCommands: info | screenshot <file> | ps <PowerShell> | quit")

        while True:
            line = input("remote> ").strip()
            if not line:
                continue

            if line == "info":
                response = await request(websocket, "SYSTEM_INFO")
                print(json.dumps(response, indent=2))
            elif line.startswith("screenshot "):
                output = Path(line.split(" ", 1)[1]).expanduser()
                response = await request(websocket, "SCREENSHOT")
                if response.get("type") != "SCREENSHOT_RESULT":
                    print(json.dumps(response, indent=2))
                    continue
                output.write_bytes(base64.b64decode(response["data"]))
                with Image.open(output) as image:
                    print(f"Saved {image.size[0]}x{image.size[1]} screenshot to {output}")
            elif line.startswith("ps "):
                response = await request(websocket, "POWERSHELL", command=line[3:].strip())
                print(json.dumps(response.get("data", response), indent=2))
            elif line == "quit":
                await request(websocket, "CLOSE")
                return
            else:
                print("Unknown command.")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Lite Remote Controller Python client")
    parser.add_argument("--host", required=True)
    parser.add_argument("--port", type=int, default=8765)
    parser.add_argument("--token", required=True)
    args = parser.parse_args()

    try:
        asyncio.run(run(args))
    except (ConnectionError, OSError, RuntimeError) as exc:
        print(f"Connection failed: {exc}", file=sys.stderr)
        raise SystemExit(1)
