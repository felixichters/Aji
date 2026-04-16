// Aji web client entry point.
//
// v0 scaffold: boots a blank PixiJS application and renders the project
// title. No networking, no board, no interaction — those land in the
// render/, state/, and net/ modules in later steps.

import { Application, Text, TextStyle } from "pixi.js";

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

  const title = new Text({
    text: "Aji (餘味)",
    style: new TextStyle({
      fill: "#f5f5f5",
      fontFamily: "ui-sans-serif, system-ui, sans-serif",
      fontSize: 48,
      fontWeight: "600",
    }),
  });
  title.anchor.set(0.5);
  title.position.set(app.screen.width / 2, app.screen.height / 2);
  app.stage.addChild(title);

  app.renderer.on("resize", (w, h) => {
    title.position.set(w / 2, h / 2);
  });
}

main().catch((err) => {
  console.error(err);
});
