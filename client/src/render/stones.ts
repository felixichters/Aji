// Stone rendering and click-to-place interaction.

import { Application, Container, Graphics } from "pixi.js";
import type { StateSnapshot } from "../net/protocol";
import { placeStone } from "../net";
import { playerColor } from "./colors";
import { CELL_SIZE, STONE_RADIUS } from "./board";
import type { Camera } from "./camera";

export class StoneRenderer {
  private container: Container;
  private stones = new Map<string, Graphics>();
  private radiusGfx: Graphics;
  private preview: Graphics;
  private camera: Camera;
  private boardW: number;
  private boardH: number;
  private radius: number;

  constructor(
    world: Container,
    camera: Camera,
    app: Application,
    boardW: number,
    boardH: number,
    radius: number,
  ) {
    this.camera = camera;
    this.boardW = boardW;
    this.boardH = boardH;
    this.radius = radius;

    this.container = new Container();
    world.addChild(this.container);

    // radius outlines drawn behind stones
    this.radiusGfx = new Graphics();
    this.container.addChild(this.radiusGfx);

    // hover preview stone
    this.preview = new Graphics();
    this.preview.visible = false;
    this.container.addChild(this.preview);

    this._bindClick(app);
    this._bindHover(app);
  }

  updateStones(snapshot: StateSnapshot, myPlayerId: string | null): void {
    // build set of current stone keys with their color + ownership
    const current = new Map<string, { color: number; own: boolean }>();
    const ownStones: { x: number; y: number }[] = [];
    for (const p of snapshot.players) {
      const color = playerColor(p.joinSeq);
      const own = p.id === myPlayerId;
      for (const c of p.stones) {
        current.set(`${c.x},${c.y}`, { color, own });
        if (own) ownStones.push(c);
      }
    }

    // find engaged player IDs
    const engagedIds = new Set<string>();
    if (myPlayerId) {
      for (const e of snapshot.engagements) {
        if (e.a === myPlayerId) engagedIds.add(e.b);
        else if (e.b === myPlayerId) engagedIds.add(e.a);
      }
    }

    // draw radius outlines
    this.radiusGfx.clear();
    const radiusPx = this.radius * CELL_SIZE;

    // engaged players' radii (dimmed)
    for (const p of snapshot.players) {
      if (!engagedIds.has(p.id)) continue;
      for (const c of p.stones) {
        this.radiusGfx.circle(c.x * CELL_SIZE, c.y * CELL_SIZE, radiusPx);
      }
      this.radiusGfx.stroke({ color: 0x222222, width: 1, alpha: 0.25 });
    }

    // own radii (on top, stronger)
    if (ownStones.length > 0) {
      for (const c of ownStones) {
        this.radiusGfx.circle(c.x * CELL_SIZE, c.y * CELL_SIZE, radiusPx);
      }
      this.radiusGfx.stroke({ color: 0x222222, width: 1.5, alpha: 0.35 });
    }

    // remove stones no longer present
    for (const [key, gfx] of this.stones) {
      if (!current.has(key)) {
        this.container.removeChild(gfx);
        gfx.destroy();
        this.stones.delete(key);
      }
    }

    // add new stones
    for (const [key, info] of current) {
      if (this.stones.has(key)) continue;

      const [sx, sy] = key.split(",").map(Number);
      const gfx = new Graphics();
      const px = sx * CELL_SIZE;
      const py = sy * CELL_SIZE;

      // stone circle
      gfx.circle(px, py, STONE_RADIUS).fill(info.color);

      // own-stone marker: small contrasting inner dot
      if (info.own) {
        const markerColor = this._contrastDot(info.color);
        gfx.circle(px, py, STONE_RADIUS * 0.25).fill(markerColor);
      }

      this.container.addChild(gfx);
      this.stones.set(key, gfx);
    }
  }

  private _contrastDot(color: number): number {
    // simple luminance check
    const r = (color >> 16) & 0xff;
    const g = (color >> 8) & 0xff;
    const b = color & 0xff;
    const lum = 0.299 * r + 0.587 * g + 0.114 * b;
    return lum > 128 ? 0x333333 : 0xeeeeee;
  }

  private _snapToGrid(sx: number, sy: number): { x: number; y: number } {
    const w = this.camera.screenToWorld(sx, sy);
    const gx = Math.round(w.x / CELL_SIZE);
    const gy = Math.round(w.y / CELL_SIZE);
    return {
      x: Math.max(0, Math.min(this.boardW - 1, gx)),
      y: Math.max(0, Math.min(this.boardH - 1, gy)),
    };
  }

  private _bindClick(app: Application): void {
    app.stage.on("pointerup", (e) => {
      if (this.camera.wasDrag()) return;
      const { x, y } = this._snapToGrid(e.global.x, e.global.y);
      placeStone(x, y);
    });
  }

  private _bindHover(app: Application): void {
    app.stage.on("pointermove", (e) => {
      if (e.buttons !== 0) {
        this.preview.visible = false;
        return;
      }
      const { x, y } = this._snapToGrid(e.global.x, e.global.y);
      const px = x * CELL_SIZE;
      const py = y * CELL_SIZE;
      this.preview.clear();
      this.preview.circle(px, py, STONE_RADIUS).fill({ color: 0x888888, alpha: 0.35 });
      this.preview.visible = true;
    });

    app.canvas.addEventListener("pointerleave", () => {
      this.preview.visible = false;
    });
  }
}
