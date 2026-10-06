# Lite Remote Controller

A lightweight educational remote-access project for connecting to a Windows machine over the internet.

## Implementations

- pythonImpl/ — original Python WebSocket prototype with basic remote commands and on-demand screenshots.
- goImpl/ — current implementation using native Go + Pion WebRTC for internet-oriented connectivity.

The Go implementation is the active path for the remote-access product.

## Current Go architecture

The project is P2P-first:
- A small public signaling server exchanges WebRTC SDP.
- The controller and Windows agent establish a WebRTC PeerConnection.
- A WebRTC DataChannel carries control messages.
- STUN helps establish direct connectivity.
- TURN will be added as a relay fallback for restrictive networks.

## Current milestone

The Go implementation currently supports:
- Internet-capable signaling
- Shared session authentication
- WebRTC DataChannel
- Remote system information
- Ping/latency test
- Remote PowerShell execution on the Windows agent
- Foreground, manually launched Windows agent

## Planned remote-control features

1. Windows screen capture
2. Live screen streaming over a WebRTC video track
3. Mouse and keyboard events over DataChannel
4. TURN fallback
5. File transfer and clipboard
6. Session lifecycle, expiry and device management
7. Native installers for Windows and macOS

## Quick test

Run the commands from goImpl/.

Start signaling:
    cd goImpl
    go run ./cmd/signaling

Create a session:
    curl http://localhost:8080/session

Start the Windows agent:
    cd goImpl
    go run .\cmd\agent --server ws://SERVER:8080/ws --session <SESSION>

Start the controller:
    cd goImpl
    go run ./cmd/controller --server ws://SERVER:8080/ws --session <SESSION>

The default STUN server is stun:stun.cloudflare.com:3478.

For public deployment, put signaling behind TLS and use wss://.

## Project boundary

This project is intended for authorized remote administration and learning. The agent is manually launched and does not implement hidden persistence, stealth, or security-software evasion.