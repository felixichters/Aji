# Aji (餘味) — a Go derivative

One large board. Many players at once. 

## Repository layout

```
server/   Go backend (fixed-grid world, per-player radius, WebSocket hub)
client/   TypeScript + PixiJS web client
docs/     Human-readable protocol and design notes
```

See `server/internal/*/doc.go` and `client/src/*` for module boundaries.

## Stack

- **Server**: Go (stdlib HTTP today; WebSocket transport lands next).
- **Client**: TypeScript + [PixiJS](https://pixijs.com) (WebGL) built with [Vite](https://vitejs.dev).
- **Protocol**: JSON over WebSocket. See `docs/protocol.md`.
- **Dev env**: Nix `flake.nix` devShell + `just` task runner.

## Quickstart

```sh
# Enter the dev shell
nix develop

# Run the Go server on :8080
just server

# In another shell, run the web client on :5173 (proxies /ws and /healthz to the server)
just client

# Or run both concurrently
just dev
```

Health check: `curl http://localhost:8080/healthz` → `ok`.

## Design choices (for v0)

| Area         | Choice                                                  |
|--------------|---------------------------------------------------------|
| Client       | Web, TypeScript + PixiJS (WebGL)                        |
| Server       | Go                                                      |
| Board        | Fixed huge grid (size configurable, e.g. 200×200)       |
| Turn model   | Per-player radius                                       |
| Persistence  | In-memory only (resets on restart)                      |

## Status

| Phase | What | State |
|-------|------|-------|
| 1 | Board logic (grid, capture, ko) | done |
| 2 | Game rules (radius, turn gating, merge) | next |
| 3 | WebSocket hub + protocol | planned |
| 4 | Client renderer (PixiJS) | planned |
| 5 | Integration | planned |
