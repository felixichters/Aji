import { Connection } from "./connection";
import type { ServerMessage } from "./protocol";
import * as gameState from "../state/gameState";

let conn: Connection | null = null;

export function connectToServer(playerId: string): void {
  const wsUrl = `ws://${location.host}/ws`;

  conn = new Connection({
    url: wsUrl,
    onOpen: () => {
      gameState.setConnected(true);
      conn!.join(playerId);
    },
    onClose: () => {
      gameState.setConnected(false);
    },
    onError: () => {
      gameState.setConnected(false);
    },
    onMessage: (msg: ServerMessage) => {
      switch (msg.type) {
        case "joined":
          gameState.applyJoined(msg.payload);
          break;
        case "state":
          gameState.applyState(msg.payload);
          break;
        case "error":
          gameState.applyError(msg.payload.code, msg.payload.message);
          break;
      }
    },
  });
  conn.connect();
}

export function placeStone(x: number, y: number): void {
  conn?.place(x, y);
}
