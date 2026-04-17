# Wire Protocol

JSON over WebSocket at `/ws`. Every frame is a JSON object with `type`
(string discriminator) and `payload` (type-specific body).

Source of truth:
- Go types: `server/internal/protocol/messages.go`
- TS types: `client/src/net/protocol.ts`
- This file: catalog

## Client -> Server

### `join`

Request to join the game.

| Field      | Type   | Description       |
|------------|--------|-------------------|
| `playerId` | string | Desired player ID |

### `place`

Place a stone.

| Field | Type | Description  |
|-------|------|--------------|
| `x`   | int  | Cell X coord |
| `y`   | int  | Cell Y coord |

## Server -> Client

### `joined`

Sent to the joining client on success.

| Field      | Type          | Description             |
|------------|---------------|-------------------------|
| `playerId` | string        | Assigned player ID      |
| `boardW`   | int           | Board width             |
| `boardH`   | int           | Board height            |
| `radius`   | int           | Region radius           |
| `state`    | StateSnapshot | Current full game state |

### `state`

Broadcast to all clients after every successful move. Body is a
`StateSnapshot` (same shape as `joined.state`).

### `error`

Sent to the acting client when their action fails.

| Field     | Type   | Description           |
|-----------|--------|-----------------------|
| `code`    | string | Machine-readable code |
| `message` | string | Human-readable detail |

Error codes: `unknown_player`, `duplicate_player`, `not_your_turn`,
`occupied`, `out_of_bounds`, `not_engaged`, `outside_region`,
`bootstrap_must_engage`, `not_joined`, `bad_request`, `internal`.

## StateSnapshot

| Field         | Type             | Description                  |
|---------------|------------------|------------------------------|
| `players`     | PlayerState[]    | All players and their stones |
| `cliques`     | CliqueState[]    | Active maximal cliques       |
| `engagements` | EngagementEdge[] | All engagement edges         |

### PlayerState

`{ id: string, joinSeq: int, stones: Cell[] }`

### Cell

`{ x: int, y: int }`

### CliqueState

`{ members: string[], toMove: string }`

### EngagementEdge

`{ a: string, b: string }`
