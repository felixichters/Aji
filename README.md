# Aji (餘味) — a Go variant

One large board. Many players at once.

## Repository layout

```
server/   Go backend (rule engine; WebSocket hub next)
client/   TypeScript + PixiJS web client (scaffold)
docs/     Design notes
```

See `server/internal/*/doc.go` and `client/src/*` for module boundaries.

## Stack

- **Server**: Go (stdlib HTTP; WebSocket transport follow-up).
- **Client**: TypeScript + [PixiJS](https://pixijs.com) (WebGL) built with [Vite](https://vitejs.dev).
- **Protocol**: JSON over WebSocket — specced once the transport implemented.
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

# Run the Go test suite
cd server && go test ./...
```

Health check: `curl http://localhost:8080/healthz` → `ok`.
