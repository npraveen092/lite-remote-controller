import argparse
import asyncio
import base64
import json
import os
import platform
import secrets
import socket
import subprocess
from io import BytesIO

import mss
from PIL import Image
from websockets.asyncio.server import serve


MAX_MESSAGE_BYTES = 2 * 1024 * 1024


def get_system_info() -> dict:
    return {
        "hostname": socket.gethostname(),
        "platform": platform.platform(),
        "python": platform.python_version(),
        "architecture": platform.machine(),
    }


def run_command(command: str) -> dict:
    # Intentionally executes only after the user has manually started this
    # agent process. No persistence or hidden execution is implemented.
    completed = subprocess.run(
        ["powershell.exe", "-NoProfile", "-NonInteractive", "-Command", command],
        capture_output=True,
        text=True,
        timeout=30,
        check=False,
    )
    return {
        "exitCode": completed.returncode,
        "stdout": completed.stdout[-20000:],
        "stderr": completed.stderr[-20000:],
    }


def capture_screen() -> str:
    with mss.mss() as sct:
        monitor = sct.monitors[1]
        shot = sct.grab(monitor)
        image = Image.frombytes("RGB", shot.size, shot.rgb)

        # JPEG keeps the control channel reasonably small for an MVP.
        buffer = BytesIO()
        image.save(buffer, format="JPEG", quality=70, optimize=True)
        return base64.b64encode(buffer.getvalue()).decode("ascii")


async def send_json(websocket, payload: dict) -> None:
    encoded = json.dumps(payload, separators=(",", ":"))
    if len(encoded.encode("utf-8")) > MAX_MESSAGE_BYTES:
        raise ValueError("Response exceeds configured message size")
    await websocket.send(encoded)


async def handle_client(websocket, args) -> None:
    authenticated = False

    try:
        raw = await asyncio.wait_for(websocket.recv(), timeout=15)
        if isinstance(raw, bytes):
            raw = raw.decode("utf-8")

        hello = json.loads(raw)
        if hello.get("type") != "AUTH" or not secrets.compare_digest(
            str(hello.get("token", "")), args.token
        ):
            await send_json(websocket, {"type": "ERROR", "error": "authentication_failed"})
            await websocket.close()
            return

        authenticated = True
        await send_json(websocket, {"type": "AUTH_OK", "system": get_system_info()})

        async for raw in websocket:
            if isinstance(raw, bytes):
                raw = raw.decode("utf-8")

            request = json.loads(raw)
            request_id = request.get("requestId")
            request_type = request.get("type")

            if request_type == "PING":
                await send_json(websocket, {"type": "PONG", "requestId": request_id})

            elif request_type == "SYSTEM_INFO":
                await send_json(
                    websocket,
                    {
                        "type": "SYSTEM_INFO_RESULT",
                        "requestId": request_id,
                        "data": get_system_info(),
                    },
                )

            elif request_type == "SCREENSHOT":
                screenshot = await asyncio.to_thread(capture_screen)
                await send_json(
                    websocket,
                    {
                        "type": "SCREENSHOT_RESULT",
                        "requestId": request_id,
                        "format": "jpeg/base64",
                        "data": screenshot,
                    },
                )

            elif request_type == "POWERSHELL":
                command = str(request.get("command", "")).strip()
                if not command:
                    await send_json(
                        websocket,
                        {
                            "type": "ERROR",
                            "requestId": request_id,
                            "error": "command_required",
                        },
                    )
                    continue

                result = await asyncio.to_thread(run_command, command)
                await send_json(
                    websocket,
                    {
                        "type": "POWERSHELL_RESULT",
                        "requestId": request_id,
                        "data": result,
                    },
                )

            elif request_type == "CLOSE":
                await websocket.close()
                return

            else:
                await send_json(
                    websocket,
                    {
                        "type": "ERROR",
                        "requestId": request_id,
                        "error": f"unsupported_request_type:{request_type}",
                    },
                )

    except asyncio.TimeoutError:
        await websocket.close()
    except (json.JSONDecodeError, TypeError, ValueError) as exc:
        await send_json(websocket, {"type": "ERROR", "error": f"invalid_message:{exc}"})
    except Exception as exc:
        if authenticated:
            try:
                await send_json(websocket, {"type": "ERROR", "error": str(exc)})
            except Exception:
                pass


async def main(host: str, port: int, token: str) -> None:
    print(f"Lite Remote Controller agent listening on ws://{host}:{port}")
    print("The agent remains active only while this process is running.")
    print(f"Session token: {token}")

    async with serve(
        lambda websocket: handle_client(websocket, argparse.Namespace(token=token)),
        host,
        port,
        max_size=MAX_MESSAGE_BYTES,
    ):
        await asyncio.Future()


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Lite Remote Controller Windows agent")
    parser.add_argument("--host", default="0.0.0.0")
    parser.add_argument("--port", type=int, default=8765)
    parser.add_argument("--token", help="Session token. A random token is generated when omitted.")
    args = parser.parse_args()

    token = args.token or secrets.token_urlsafe(18)
    try:
        asyncio.run(main(args.host, args.port, token))
    except KeyboardInterrupt:
        print("\nAgent stopped.")
