import type { StateSnapshot, JoinedMsg } from "../net/protocol";

export interface GameState {
  connected: boolean;
  joined: boolean;
  playerId: string | null;
  boardW: number;
  boardH: number;
  radius: number;
  snapshot: StateSnapshot | null;
  lastError: { code: string; message: string } | null;
}

export type Subscriber = (state: Readonly<GameState>) => void;

const initial: GameState = {
  connected: false,
  joined: false,
  playerId: null,
  boardW: 0,
  boardH: 0,
  radius: 0,
  snapshot: null,
  lastError: null,
};

let state: GameState = { ...initial };
const subscribers: Set<Subscriber> = new Set();

function notify(): void {
  for (const fn of subscribers) fn(state);
}

export function subscribe(fn: Subscriber): () => void {
  subscribers.add(fn);
  fn(state);
  return () => subscribers.delete(fn);
}

export function getState(): Readonly<GameState> {
  return state;
}

export function setConnected(connected: boolean): void {
  state = { ...state, connected };
  if (!connected) state = { ...state, joined: false };
  notify();
}

export function applyJoined(msg: JoinedMsg): void {
  state = {
    ...state,
    joined: true,
    playerId: msg.playerId,
    boardW: msg.boardW,
    boardH: msg.boardH,
    radius: msg.radius,
    snapshot: msg.state,
    lastError: null,
  };
  notify();
}

export function applyState(snapshot: StateSnapshot): void {
  state = { ...state, snapshot, lastError: null };
  notify();
}

export function applyError(code: string, message: string): void {
  state = { ...state, lastError: { code, message } };
  notify();
}
