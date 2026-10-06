# Architecture

## Components

### Controller

Runs on the operator machine. It is responsible for:

- Creating/initiating a session
- Authenticating to the agent
- Receiving responses
- Later rendering remote screen frames
- Later sending input events

### Windows Agent

Runs only after being explicitly launched by the user.

Initial responsibilities:

- Generate a session identity
- Establish an authenticated connection
- Execute approved commands
- Return structured responses
- Provide a future extension point for screen capture

## Protocol

Messages should be versioned and explicit. A future message envelope can look like:

```json
{
  "version": 1,
  "type": "COMMAND",
  "requestId": "uuid",
  "payload": {}
}
```

Potential message types:

- HELLO
- AUTH
- COMMAND
- COMMAND_RESULT
- SCREEN_FRAME
- INPUT_EVENT
- ERROR
- CLOSE

## Transport

Start with a local/LAN transport so behavior can be tested without NAT traversal.

Internet connectivity can later use TLS plus a signaling/relay design.

## Design principles

- Explicit session lifecycle
- Strong authentication
- Encryption in transit
- Minimal privileges
- Auditable actions
- No stealth or evasion mechanisms
