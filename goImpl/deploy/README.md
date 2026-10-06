# VPS Deployment

This directory contains the minimal public deployment for the Go implementation.

## Services

The VPS runs:
- signaling service
- coturn TURN relay
- Caddy TLS reverse proxy

A single small Linux VPS is sufficient for the first deployment.

## 1. Prepare the VPS

Install Docker Engine and Docker Compose.
Create a directory for the project and clone this repository.

## 2. Configure DNS

Create an A record such as:
    signal.example.com -> <VPS_PUBLIC_IPV4>

Use your own domain and hostname.

## 3. Configure environment

Copy:
    cp .env.example .env

Edit .env and set:
- SIGNALING_DOMAIN
- TURN_REALM
- TURN_USERNAME
- TURN_PASSWORD

Generate a long random TURN password and keep .env private.

## 4. Configure Caddy

Replace the placeholder hostname in Caddyfile with your actual signaling hostname.
Caddy will request and renew the TLS certificate automatically when ports 80 and 443 are publicly reachable.

## 5. Open the VPS firewall

TCP: 80, 443, 3478
UDP: 3478, 49160-49260

Do not expose the signaling container's internal port 8080 directly to the internet.

## 6. Start

    docker compose --env-file .env up -d --build

Verify:
    curl https://signal.example.com/healthz

It should return:
    ok

Generate a session:
    curl https://signal.example.com/session

## 7. Start the Windows agent

The command-line parameters support STUN and TURN:

    go run .\cmd\agent --server wss://signal.example.com/ws --session <SESSION> --stun stun:stun.cloudflare.com:3478 --turn turn:signal.example.com:3478 --turn-user <TURN_USERNAME> --turn-credential <TURN_PASSWORD>

For a deployment build, use a compiled Windows executable instead of go run.

## 8. Start the controller

    go run ./cmd/controller --server wss://signal.example.com/ws --session <SESSION> --stun stun:stun.cloudflare.com:3478 --turn turn:signal.example.com:3478 --turn-user <TURN_USERNAME> --turn-credential <TURN_PASSWORD>

## Important

The signaling URL should be wss:// in production.
The project currently provides the internet transport and TURN-ready configuration. The remote screen/input implementation remains a separate component.
