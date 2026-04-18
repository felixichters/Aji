// Pan & zoom camera backed by a PixiJS Container transform.

import { Application, Container, Rectangle } from "pixi.js";

const MIN_ZOOM = 0.05;
const MAX_ZOOM = 4.0;
const DRAG_THRESHOLD = 3; // px before a pointerdown counts as drag

export interface ViewportBounds {
  minX: number;
  minY: number;
  maxX: number;
  maxY: number;
}

export class Camera {
  readonly world: Container;
  private app: Application;
  private _dirty = true;
  private _wasDrag = false;

  // drag state
  private _dragging = false;
  private _dragStartX = 0;
  private _dragStartY = 0;
  private _dragDist = 0;
  private _lastPointerX = 0;
  private _lastPointerY = 0;

  constructor(app: Application) {
    this.app = app;
    this.world = new Container();
    app.stage.addChild(this.world);

    // make stage interactive for pan events
    app.stage.eventMode = "static";
    app.stage.hitArea = new Rectangle(0, 0, app.screen.width, app.screen.height);

    app.renderer.on("resize", (w: number, h: number) => {
      app.stage.hitArea = new Rectangle(0, 0, w, h);
      this._dirty = true;
    });

    this._bindPan();
    this._bindZoom();
  }

  /** Center the board and fit it in the viewport. */
  centerOn(boardW: number, boardH: number, cellSize: number): void {
    const worldW = (boardW - 1) * cellSize;
    const worldH = (boardH - 1) * cellSize;
    const scaleX = this.app.screen.width / (worldW + cellSize * 2);
    const scaleY = this.app.screen.height / (worldH + cellSize * 2);
    const scale = Math.min(scaleX, scaleY, MAX_ZOOM);

    this.world.scale.set(scale);
    this.world.position.set(
      (this.app.screen.width - worldW * scale) / 2,
      (this.app.screen.height - worldH * scale) / 2,
    );
    this._dirty = true;
  }

  /** True if the camera transform changed since last resetDirty(). */
  get dirty(): boolean {
    return this._dirty;
  }

  resetDirty(): void {
    this._dirty = false;
  }

  /** True if the last pointer interaction was a drag (not a click). */
  wasDrag(): boolean {
    return this._wasDrag;
  }

  getViewportBounds(): ViewportBounds {
    const s = this.world.scale.x;
    const minX = -this.world.position.x / s;
    const minY = -this.world.position.y / s;
    return {
      minX,
      minY,
      maxX: minX + this.app.screen.width / s,
      maxY: minY + this.app.screen.height / s,
    };
  }

  /** Convert screen coordinates to world coordinates. */
  screenToWorld(sx: number, sy: number): { x: number; y: number } {
    const s = this.world.scale.x;
    return {
      x: (sx - this.world.position.x) / s,
      y: (sy - this.world.position.y) / s,
    };
  }

  // --- internals ---

  private _bindPan(): void {
    const stage = this.app.stage;

    stage.on("pointerdown", (e) => {
      this._dragging = false;
      this._wasDrag = false;
      this._dragStartX = e.global.x;
      this._dragStartY = e.global.y;
      this._lastPointerX = e.global.x;
      this._lastPointerY = e.global.y;
      this._dragDist = 0;
    });

    stage.on("pointermove", (e) => {
      if (this._dragStartX === 0 && this._dragStartY === 0 && !this._dragging) return;
      // only count moves while a button is pressed
      if (e.buttons === 0) return;

      const dx = e.global.x - this._lastPointerX;
      const dy = e.global.y - this._lastPointerY;
      this._lastPointerX = e.global.x;
      this._lastPointerY = e.global.y;

      this._dragDist += Math.abs(e.global.x - this._dragStartX) + Math.abs(e.global.y - this._dragStartY);

      if (!this._dragging && this._dragDist > DRAG_THRESHOLD) {
        this._dragging = true;
      }

      if (this._dragging) {
        this.world.position.x += dx;
        this.world.position.y += dy;
        this._dirty = true;
      }
    });

    stage.on("pointerup", () => {
      if (this._dragging) {
        this._wasDrag = true;
      }
      this._dragging = false;
    });

    stage.on("pointerupoutside", () => {
      this._dragging = false;
    });
  }

  private _bindZoom(): void {
    this.app.canvas.addEventListener("wheel", (e: WheelEvent) => {
      e.preventDefault();

      const oldScale = this.world.scale.x;
      const factor = e.deltaY < 0 ? 1.1 : 1 / 1.1;
      const newScale = Math.min(MAX_ZOOM, Math.max(MIN_ZOOM, oldScale * factor));

      // zoom toward pointer position
      const rect = this.app.canvas.getBoundingClientRect();
      const px = e.clientX - rect.left;
      const py = e.clientY - rect.top;

      const worldX = (px - this.world.position.x) / oldScale;
      const worldY = (py - this.world.position.y) / oldScale;

      this.world.scale.set(newScale);
      this.world.position.set(px - worldX * newScale, py - worldY * newScale);

      this._dirty = true;
    }, { passive: false });
  }
}
