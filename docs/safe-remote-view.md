# Safe Remote View

This repository includes a transport/test harness for authenticated remote viewing.

## Authorization

The signaling server authenticates both peers with the per-session bearer token. The token is sent in the `Authorization` header and is not placed in the WebSocket URL.

## Screen transport

`goImpl/internal/safeview/frame.go` defines a small frame-chunk protocol for carrying JPEG/PNG or other encoded frame bytes over a WebRTC data channel. It deliberately does not implement Windows desktop capture or OS-level input injection.

## Input harness

`goImpl/internal/safeview/input.go` parses/receives input events and logs them. It is intended for protocol and integration testing without changing the host's mouse or keyboard state.

## Next safe integration

An application UI can consume decoded frame bytes for display and send `InputEvent` messages for end-to-end protocol testing. The host-side handler remains non-injecting.