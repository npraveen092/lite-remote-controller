# Remote Control / Screen Capture Development Prompt

Use this prompt as the implementation brief for the next development phase of `lite-remote-controller`.

## Project goal

Build a production-quality, token-authorized remote-support application for systems owned or explicitly administered by the operator.

Current transport:
- Go 1.24
- Pion WebRTC
- Gorilla WebSocket
- HTTPS/WSS signaling
- STUN/TURN
- Per-session bearer token

## Required architecture

```text
Controller
   |
   | HTTPS/WSS signaling + bearer token
   v
Signaling Server
   |
   | SDP offer/answer
   v
WebRTC
   |--------------------------|
   |                          |
Screen media/data         Control events
   |                          |
   v                          v
Agent                         Agent
```

## Screen capture implementation

Implement the Windows desktop capture layer for the authenticated agent.

Requirements:
- Capture the primary desktop at a configurable FPS.
- Encode frames efficiently, preferably JPEG for an initial prototype.
- Include frame ID, timestamp, dimensions, encoding, and sequence metadata.
- Fragment large frames according to the existing `safeview` chunk protocol.
- Send frames over the established authenticated WebRTC connection.
- Handle dropped, delayed, duplicated, and out-of-order chunks.
- Reassemble frames on the controller.
- Render the latest complete frame in a controller UI.
- Avoid unbounded memory growth when frames arrive faster than they can be rendered.
- Add metrics for FPS, frame size, latency, dropped frames, and reconnects.

## Session authorization

- Require a high-entropy per-session token.
- Send the token only through the HTTPS/WSS Authorization header.
- Never put the token in a URL.
- Bind the token to exactly one agent and one controller session.
- Expire or invalidate sessions when the session ends.
- Reject additional peers after the session is occupied.
- Do not expose the signaling server directly to the public Internet when a TLS reverse proxy is configured.

## Control-channel development

Implement a typed control-event protocol with explicit session authorization and lifecycle handling.

Supported event categories for the test harness:
- mouse move
- mouse button press/release
- keyboard key press/release
- modifier state
- controller disconnect
- session close

Validate every event before processing:
- message type
- request ID
- payload size
- coordinate range
- key/event allowlist where applicable
- session state

During development, keep the host-side input handler as a non-injecting test harness so the network/control protocol can be validated without changing the host OS input state.

## Reliability

Add:
- reconnect handling
- heartbeat/timeout handling
- clean WebRTC shutdown
- controller and agent logs
- session diagnostics
- TURN fallback testing
- automated unit tests for frame fragmentation/reassembly
- integration tests for token authentication

## Deliverables

1. Windows agent screen-capture module.
2. WebRTC screen transport.
3. Controller frame reassembly and rendering.
4. Authenticated control-event protocol.
5. Integration and unit tests.
6. VPS deployment documentation.
7. Performance and troubleshooting documentation.

## Security requirements

Treat the bearer token as a session credential. Never hard-code real credentials or commit secrets. Keep the application foreground-visible and intended for machines the operator is authorized to administer.

Do not add persistence, stealth, security-tool bypasses, credential theft, or hidden/background execution.