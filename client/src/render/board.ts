// Go board grid renderer with viewport culling.

import { Container, Graphics } from "pixi.js";
import type { ViewportBounds } from "./camera";

export const CELL_SIZE = 30;
export const STONE_RADIUS = CELL_SIZE * 0.45;

const BOARD_COLOR = 0xd2a259;
const LINE_COLOR = 0x333333;
const LINE_WIDTH = 1;

export class BoardRenderer {
  private gfx: Graphics;
  private boardW: number;
  private boardH: number;

  constructor(world: Container, boardW: number, boardH: number) {
    this.boardW = boardW;
    this.boardH = boardH;
    this.gfx = new Graphics();
    world.addChild(this.gfx);
  }

  updateVisibleArea(bounds: ViewportBounds): void {
    const { boardW, boardH } = this;
    const margin = CELL_SIZE * 2;

    const startCol = Math.max(0, Math.floor((bounds.minX - margin) / CELL_SIZE));
    const endCol = Math.min(boardW - 1, Math.ceil((bounds.maxX + margin) / CELL_SIZE));
    const startRow = Math.max(0, Math.floor((bounds.minY - margin) / CELL_SIZE));
    const endRow = Math.min(boardH - 1, Math.ceil((bounds.maxY + margin) / CELL_SIZE));

    const g = this.gfx;
    g.clear();

    // wooden background covering the full board
    const pad = CELL_SIZE / 2;
    g.rect(-pad, -pad, (boardW - 1) * CELL_SIZE + pad * 2, (boardH - 1) * CELL_SIZE + pad * 2);
    g.fill(BOARD_COLOR);

    // vertical lines (columns)
    const yTop = startRow * CELL_SIZE;
    const yBot = endRow * CELL_SIZE;
    for (let col = startCol; col <= endCol; col++) {
      const x = col * CELL_SIZE;
      g.moveTo(x, yTop).lineTo(x, yBot);
    }

    // horizontal lines (rows)
    const xLeft = startCol * CELL_SIZE;
    const xRight = endCol * CELL_SIZE;
    for (let row = startRow; row <= endRow; row++) {
      const y = row * CELL_SIZE;
      g.moveTo(xLeft, y).lineTo(xRight, y);
    }

    g.stroke({ color: LINE_COLOR, width: LINE_WIDTH });
  }
}
