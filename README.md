# Lite Remote Controller

A lightweight, educational remote-access tool for connecting to a Windows machine over the internet.

## Current architecture

The project is P2P-first:

- A small public signaling server exchanges WebRTC SDP.
- The controller and Windows agent establish a WebRTC PeerConnection.
- A WebRTC DataChannel carries control messages.
- STUN helps establish direct connectivity.
- TURN will be added as a relay fallback for restrictive networks.

Pion is used for the native WebRTC layer. Pion is a pure-Go WebRTC implementation. https://github.com/pion/webrtc

## Current milestone

The repository currently supports:

- Internet-capable signaling
- Shared session authentication
- WebRTC DataChannel
- Remote system information
- Ping/latency test
- Remote PowerShell execution on the Windows agent
- Foreground, manually launched Windows agent

## Planned remote-control features

1. Screen capture
2. Live screen streaming over a WebRTC video track
3. Mouse and keyboard events over DataChannel
4. TURN fallback
5. File transfer and clipboard
6. Session lifecycle, expiry and device management
7. Native installers for Windows and macOS

## Quick test

Start the signaling service:

    go run ./cmd/signaling

Create a session token:

    curl http://localhost:8080/session

Start the Windows agent:

    go run .\cmd\agent --server ws://SERVER:8080/ws --session <SESSION>

Start the controller:

    go run ./cmd/controller --server ws://SERVER:8080/ws --session <SESSION>

The default STUN server is stun.cloudflare.com:3478.

For public deployment, put signaling behind TLS and use wss://.

## Project boundary

This project is intended for authorized remote administration and learning. The agent is manually launched and does not implement hidden persistence, stealth, or security-software evasion.
