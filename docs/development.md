# Development

## Implementations

### Python prototype

The original LAN-oriented prototype remains under pythonImpl/.
It is useful for experimentation with screen capture and the initial WebSocket protocol.

### Go/WebRTC implementation

The active implementation is under goImpl/.

Requirements:
- Go 1.24+
- A Windows test machine for the agent
- A Mac/Linux/Windows machine for the controller
- A reachable signaling server for internet testing

## Run signaling locally
    cd goImpl
    go run ./cmd/signaling

## Generate a test session
    curl http://localhost:8080/session

Use the returned random token for both peers.

## Run the Windows agent
    cd goImpl
    go run .\cmd\agent --server ws://SERVER:8080/ws --session <SESSION>

The agent is a foreground process. Stop it with Ctrl+C.

## Run the controller
    cd goImpl
    go run ./cmd/controller --server ws://SERVER:8080/ws --session <SESSION>

The default STUN server is stun.cloudflare.com:3478.

For a local-only test with no STUN discovery, pass --stun empty.

## Current control commands
    info
    ping
    ps Get-Process
    ps Get-Service
    quit

## Public deployment

Put the signaling server behind TLS and use wss:// in the controller/agent arguments.

The current Go release is a networking foundation. It has:
- WebRTC DataChannel transport
- Shared-session authentication
- P2P-first ICE configuration
- Small standalone signaling service

It does not yet have:
- TURN fallback
- Screen streaming
- Mouse or keyboard input
- File transfer
- Device persistence