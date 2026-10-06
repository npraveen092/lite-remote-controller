# Go Implementation

This is the active remote-access implementation.

## Components

- cmd/agent — Windows agent
- cmd/controller — controller
- cmd/signaling — WebRTC signaling server
- internal/protocol — shared protocol types
- deploy — signaling deployment files

The transport is native Go + Pion WebRTC, with STUN for P2P connectivity and TURN planned as fallback.