import { Application } from "pixi.js";
import { connectToServer } from "./net";
import { initRenderer } from "./render";

async function main(): Promise<void> {
  const host = document.getElementById("app");
  if (!host) throw new Error("missing #app container");

  const app = new Application();
  await app.init({
    resizeTo: window,
    background: "#1a1a1a",
    antialias: true,
  });
  host.appendChild(app.canvas);

  const playerId = "player-" + Math.random().toString(36).slice(2, 8);
  connectToServer(playerId);
  initRenderer(app);
}

main().catch((err) => {
  console.error(err);
});
