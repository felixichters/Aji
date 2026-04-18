// Side panel: turn indicator, error flash, clique & engagement info.

import type { GameState } from "../state/gameState";
import { playerColorCSS } from "../render/colors";

const PANEL_WIDTH = 220;

let panel: HTMLElement;
let turnEl: HTMLElement;
let errorEl: HTMLElement;
let cliquesEl: HTMLElement;
let errorTimer: ReturnType<typeof setTimeout> | null = null;
let lastErrorKey: string | null = null;

export function createPanel(): void {
  const style = document.createElement("style");
  style.textContent = `
    #aji-panel {
      position: fixed;
      right: 0; top: 0; bottom: 0;
      width: ${PANEL_WIDTH}px;
      background: rgba(0, 0, 0, 0.75);
      color: #eee;
      font-family: ui-sans-serif, system-ui, sans-serif;
      font-size: 13px;
      padding: 16px 14px;
      box-sizing: border-box;
      overflow-y: auto;
      z-index: 10;
      display: flex;
      flex-direction: column;
      gap: 14px;
    }
    #aji-panel h3 {
      margin: 0 0 6px 0;
      font-size: 11px;
      text-transform: uppercase;
      letter-spacing: 0.05em;
      color: #888;
    }
    #aji-turn {
      font-size: 15px;
      font-weight: 600;
      padding: 8px 10px;
      border-radius: 6px;
      background: rgba(255,255,255,0.06);
    }
    #aji-turn.your-turn { color: #66dd88; }
    #aji-turn.waiting   { color: #999; }
    #aji-turn.solo      { color: #777; }
    #aji-error {
      padding: 6px 10px;
      border-radius: 6px;
      background: rgba(220, 60, 60, 0.25);
      color: #ff9999;
      font-size: 12px;
      opacity: 0;
      transition: opacity 0.3s;
      min-height: 0;
      overflow: hidden;
    }
    #aji-error.visible { opacity: 1; }
    .aji-clique {
      padding: 6px 8px;
      border-radius: 4px;
      background: rgba(255,255,255,0.04);
      margin-bottom: 4px;
    }
    .aji-player-dot {
      display: inline-block;
      width: 10px; height: 10px;
      border-radius: 50%;
      margin-right: 4px;
      vertical-align: middle;
      border: 1px solid rgba(255,255,255,0.2);
    }
    .aji-to-move { font-weight: 600; }
  `;
  document.head.appendChild(style);

  panel = document.createElement("div");
  panel.id = "aji-panel";

  turnEl = document.createElement("div");
  turnEl.id = "aji-turn";
  turnEl.className = "solo";
  turnEl.textContent = "Connecting\u2026";
  panel.appendChild(turnEl);

  errorEl = document.createElement("div");
  errorEl.id = "aji-error";
  panel.appendChild(errorEl);

  cliquesEl = document.createElement("div");
  panel.appendChild(cliquesEl);

  document.body.appendChild(panel);
}

export function updatePanel(state: GameState): void {
  if (!panel) return;

  // build player lookup: id -> joinSeq
  const playerSeq = new Map<string, number>();
  if (state.snapshot) {
    for (const p of state.snapshot.players) {
      playerSeq.set(p.id, p.joinSeq);
    }
  }

  // --- turn status ---
  if (!state.connected) {
    turnEl.textContent = "Disconnected";
    turnEl.className = "solo";
  } else if (!state.joined) {
    turnEl.textContent = "Connecting\u2026";
    turnEl.className = "solo";
  } else if (!state.snapshot) {
    turnEl.textContent = "Waiting for state\u2026";
    turnEl.className = "solo";
  } else {
    const myCliques = state.snapshot.cliques.filter(
      (c) => c.members.includes(state.playerId!),
    );
    if (myCliques.length === 0) {
      turnEl.textContent = "Not engaged";
      turnEl.className = "solo";
    } else {
      const isMyTurn = myCliques.every((c) => c.toMove === state.playerId);
      if (isMyTurn) {
        turnEl.textContent = "Your turn";
        turnEl.className = "your-turn";
      } else {
        const waitingOn = myCliques
          .filter((c) => c.toMove !== state.playerId)
          .map((c) => c.toMove);
        const unique = [...new Set(waitingOn)];
        turnEl.textContent = `Waiting for ${unique.join(", ")}`;
        turnEl.className = "waiting";
      }
    }
  }

  // --- error flash ---
  if (state.lastError) {
    const key = `${state.lastError.code}:${state.lastError.message}`;
    if (key !== lastErrorKey) {
      lastErrorKey = key;
      errorEl.textContent = state.lastError.message;
      errorEl.classList.add("visible");
      if (errorTimer) clearTimeout(errorTimer);
      errorTimer = setTimeout(() => {
        errorEl.classList.remove("visible");
      }, 3000);
    }
  }

  // --- cliques & engagements ---
  if (!state.snapshot || !state.playerId) {
    cliquesEl.innerHTML = "";
    return;
  }

  const myCliques = state.snapshot.cliques.filter(
    (c) => c.members.includes(state.playerId!),
  );
  const myEngagements = state.snapshot.engagements.filter(
    (e) => e.a === state.playerId || e.b === state.playerId,
  );

  let html = "";

  if (myCliques.length > 0) {
    html += "<h3>Cliques</h3>";
    for (const clique of myCliques) {
      html += '<div class="aji-clique">';
      for (const m of clique.members) {
        const seq = playerSeq.get(m) ?? 0;
        const css = playerColorCSS(seq);
        const isTurn = m === clique.toMove;
        const cls = isTurn ? "aji-to-move" : "";
        const label = m === state.playerId ? "you" : m;
        html += `<span class="${cls}"><span class="aji-player-dot" style="background:${css}"></span>${label}</span> `;
      }
      html += "</div>";
    }
  }

  if (myEngagements.length > 0) {
    html += "<h3>Engagements</h3>";
    for (const e of myEngagements) {
      const other = e.a === state.playerId ? e.b : e.a;
      const seq = playerSeq.get(other) ?? 0;
      const css = playerColorCSS(seq);
      html += `<div><span class="aji-player-dot" style="background:${css}"></span>${other}</div>`;
    }
  }

  cliquesEl.innerHTML = html;
}
