# Lite Remote Controller

An educational remote-support/control project for Windows.

## Scope

The first milestone is an explicitly user-launched remote session:

- Windows agent is started manually from a shell.
- The operator connects using a session identifier.
- The session supports authenticated communication.
- Later milestones may add screen streaming and remote input.
- No persistence, stealth, security-software evasion, or hidden remote access is part of this project.

## Planned architecture

```
+----------------+             +------------------+
| Controller     |  encrypted  | Windows Agent    |
| macOS/Windows  |<----------->|                 |
+----------------+   session   +------------------+
                                      |
                                      +-- system info
                                      +-- remote shell
                                      +-- screen capture
                                      +-- input (later)
```

## Roadmap

1. Repository and protocol skeleton
2. Local authenticated session
3. Remote command execution
4. Screenshot capture
5. Screen streaming
6. Mouse/keyboard control
7. Internet transport and NAT traversal
8. Packaging and test automation

## Security

This repository is intended for authorized testing and learning. Remote sessions should remain user-launched and observable. Do not use the project to hide activity, bypass security controls, or access systems without authorization.
