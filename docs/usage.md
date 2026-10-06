# MVP Usage

This first implementation is intentionally local/LAN-oriented.

## Windows agent

Install Python 3.11+ and dependencies:

```powershell
cd agent
python -m venv .venv
.\.venv\Scripts\Activate.ps1
pip install -r requirements.txt
python agent.py
```

The agent prints a random session token. Keep it private.

## Controller

On the Mac/Linux/Windows controller:

```bash
cd controller
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
python controller.py --host <WINDOWS_IP> --port 8765 --token <TOKEN>
```

On Windows, activate with:

```powershell
.\.venv\Scripts\Activate.ps1
```

## Available MVP commands

```text
info
screenshot /tmp/remote.jpg
ps Get-Process
ps Get-Service
quit
```

The agent must be manually launched and remains active only while its process is running. Do not expose this prototype directly to the public internet; the next phase should add TLS, proper authentication, and an authenticated relay/NAT-traversal design.
