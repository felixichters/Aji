# CLAUDE.md

Project-specific guidance for Claude Code working in this repo. Keep it
concise; for human-facing docs see `README.md`.

## What this project is

Aji is a multiplayer Go/Baduk variant: one shared board, many
players, **per-player radius** decides whose turn it is locally. Local
games drift and merge as players move.

Status: **v0** — the rule engine, WebSocket hub, wire protocol, and
client connection + state store are implemented. Client board rendering
and captures/ko are **not** written yet. See `docs/turn-rules.md` for
the rule spec and `docs/protocol.md` for the wire protocol.

## Repo layout (high level)

```
server/   Go backend
client/   TypeScript + PixiJS web client
docs/     Human-readable design / protocol notes
```

Source-of-truth references:
- Module boundaries: each `server/internal/*/doc.go` states the package's
  responsibility. **Respect them** — e.g. `board` must not import
  `player`; only `net` may touch sockets.
- Turn & region rules: `docs/turn-rules.md` is the human contract for
  the engine in `server/internal/game`. If the code and the doc
  disagree, the code wins — update the doc in the same change.
- Wire protocol: `docs/protocol.md` is the human catalog. Changes to
  message shapes must land in `server/internal/protocol/messages.go`,
  `client/src/net/protocol.ts`, and `docs/protocol.md` in the same
  change.

## Stack

| Layer      | Choice                                    |
|------------|-------------------------------------------|
| Server     | Go (stdlib HTTP today)                    |
| Client     | TypeScript + PixiJS v8 (WebGL) via Vite   |
| Transport  | JSON over WebSocket at `/ws`              |
| Persistence| In-memory only for v0                     |
| Dev env    | Nix `flake.nix` devShell + `just`         |

## Running things (NixOS — no globally installed toolchains)

Prefer `just` targets over invoking tools directly:

```sh
just server   # go run ./cmd/aji-server — listens on :8080
just client   # pnpm install + vite dev on :5173
just dev      # both concurrently
just check    # go build ./... + tsc --noEmit
just fmt      # gofmt + (future) prettier
```

If `just` isn't on PATH, wrap with `nix-shell`:

```sh
nix-shell -p go --run "go build ./..."                      # in server/
nix-shell -p nodejs_20 pnpm --run "pnpm run typecheck"      # in client/
```

Do **not** call `python`, `node`, `go`, `pnpm` bare from Bash —
they're only available inside the devShell or via `nix-shell -p`.

## v0 scope constraints (locked with the user)

- Fixed huge grid (configurable size, start ~200×200).
- Per-player radius for turn gating.
- Stones only — **no obstacles, no shrinking zone, no accounts**.
- Single server process, in-memory world.

Do not expand scope beyond what the current task asks for. Obstacles,
battle-royale zone, chunked/infinite boards, replays, and horizontal
scaling are explicitly deferred — see `Out of scope` in the latest plan
under `~/.claude/plans/`.

## Conventions

- **No speculative code.** Package boundaries exist (the `doc.go`
  files); don't fill them with scaffolding until a task actually needs
  that code.
- **Protocol changes are three-way.** Any edit to message shapes must
  land in `server/internal/protocol`, `client/src/net/protocol.ts`, AND
  `docs/protocol.md` in the same change.
- **Server is authoritative.** The client mirrors state it receives;
  never let client-side rule checks substitute for server-side ones.
- **Keep `board` pure.** It must not know about players, turns, or
  sockets. Rule novelty (radius, merging) lives in `game`.
- **Go module path** is `github.com/felixichters/Aji/server` (matches
  `server/go.mod`) — placeholder; if the user sets up a real remote,
  update `go.mod` and all imports.

## Quick verification after changes

- Server: `go build ./... && go vet ./... && go test ./internal/...`
  from `server/`.
- Client: `pnpm run typecheck` from `client/`.
- Health probe: `curl http://localhost:8080/healthz` → `ok`.