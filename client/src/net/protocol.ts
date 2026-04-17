// Wire protocol types. Must stay in sync with
// server/internal/protocol/messages.go and docs/protocol.md.

// --- Envelope ---

export interface Envelope {
  type: string;
  payload: unknown;
}

// --- Client -> Server ---

export interface JoinMsg {
  playerId: string;
}

export interface PlaceMsg {
  x: number;
  y: number;
}

// --- Server -> Client ---

export interface JoinedMsg {
  playerId: string;
  boardW: number;
  boardH: number;
  radius: number;
  state: StateSnapshot;
}

export interface StateSnapshot {
  players: PlayerState[];
  cliques: CliqueState[];
  engagements: EngagementEdge[];
}

export interface PlayerState {
  id: string;
  joinSeq: number;
  stones: Cell[];
}

export interface Cell {
  x: number;
  y: number;
}

export interface CliqueState {
  members: string[];
  toMove: string;
}

export interface EngagementEdge {
  a: string;
  b: string;
}

export interface ErrorMsg {
  code: string;
  message: string;
}

// --- Discriminated incoming messages ---

export type ServerMessage =
  | { type: "joined"; payload: JoinedMsg }
  | { type: "state"; payload: StateSnapshot }
  | { type: "error"; payload: ErrorMsg };

// --- Helpers ---

export function encodeJoin(playerId: string): string {
  return JSON.stringify({ type: "join", payload: { playerId } });
}

export function encodePlace(x: number, y: number): string {
  return JSON.stringify({ type: "place", payload: { x, y } });
}

export function decodeServerMessage(raw: string): ServerMessage {
  return JSON.parse(raw) as ServerMessage;
}
