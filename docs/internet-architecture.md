# Internet Architecture

The target architecture is a P2P-first remote-access tool.

The public signaling service only coordinates session negotiation. The preferred data path is WebRTC directly between the controller and Windows agent. TURN is a fallback for networks where direct connectivity cannot be established.

## Connection

Internet

Controller <---- WebRTC ----> Windows Agent
     \                 /
      \               /
       Signaling server

The signaling server exchanges the SDP offer and answer. WebRTC ICE then chooses a working network path.

## NAT traversal

The first internet milestone uses STUN. Cloudflare currently documents stun.cloudflare.com:3478 as a public STUN endpoint and states that its STUN service is free and unlimited. TURN is the next milestone for restrictive NAT/firewall cases. 

## TURN

TURN becomes necessary when a direct path is not possible. We can self-host coturn on a small public VM, or use a managed TURN provider. Coturn is open source and provides STUN/TURN server functionality. citeturn256343search1

## Security

For public deployment:
- Use WSS/HTTPS for signaling.
- Treat the session token as a credential.
- Use TLS certificates on the signaling endpoint.
- Do not expose a raw Windows control port to the internet.
- Add rate limiting and session expiry before broader use.
- Keep the Windows agent manually launched until the core protocol is stable.
