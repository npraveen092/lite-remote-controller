#!/usr/bin/env sh
set -eu

: "${SIGNALING_DOMAIN:?SIGNALING_DOMAIN is required}"
: "${TURN_REALM:?TURN_REALM is required}"
: "${TURN_USERNAME:?TURN_USERNAME is required}"
: "${TURN_PASSWORD:?TURN_PASSWORD is required}"

docker compose --env-file .env config >/dev/null

echo "Deployment configuration is valid."
