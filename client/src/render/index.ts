// Render subsystem init — wires camera, board, stones, and HUD panel.

import { Application } from "pixi.js";
import { Camera } from "./camera";
import { BoardRenderer } from "./board";
import { StoneRenderer } from "./stones";
import { subscribe } from "../state/gameState";
import { createPanel, updatePanel } from "../ui/panel";

export function initRenderer(app: Application): void {
  let camera: Camera | null = null;
  let board: BoardRenderer | null = null;
  let stones: StoneRenderer | null = null;

  createPanel();

  subscribe((state) => {
    updatePanel(state);

    if (!state.joined || state.boardW === 0) return;

    // first join — create renderers
    if (!camera) {
      camera = new Camera(app);
      board = new BoardRenderer(camera.world, state.boardW, state.boardH);
      stones = new StoneRenderer(camera.world, camera, app, state.boardW, state.boardH, state.radius);
      camera.centerOn(state.boardW, state.boardH, 30);
    }

    // update stones on every state change
    if (state.snapshot && stones) {
      stones.updateStones(state.snapshot, state.playerId);
    }
  });

  // per-frame: redraw grid when camera moves
  app.ticker.add(() => {
    if (camera?.dirty && board) {
      board.updateVisibleArea(camera.getViewportBounds());
      camera.resetDirty();
    }
  });
}
