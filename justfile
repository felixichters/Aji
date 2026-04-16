# Aji

default:
    @just --list

# Run the Go server on :8080
server:
    cd server && go run ./cmd/aji-server

# Run the Vite dev server for the web client on :5173
client:
    cd client && pnpm install --silent && pnpm run dev

# Run server + client concurrently (needs a shell with `&` support)
dev:
    #!/usr/bin/env bash
    set -euo pipefail
    trap 'kill 0' EXIT
    (cd server && go run ./cmd/aji-server) &
    (cd client && pnpm install --silent && pnpm run dev) &
    wait

# Format Go and TypeScript sources
fmt:
    cd server && go fmt ./...
    cd client && pnpm run fmt || true

# Type-check / build check without running
check:
    cd server && go build ./...
    cd client && pnpm install --silent && pnpm run typecheck
