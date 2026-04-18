// Deterministic player color assignment from join order.

const PALETTE: number[] = [
  0x222222, // near-black
  0xe0e0e0, // near-white
  0xe05050, // red
  0x4488ee, // blue
  0x44bb44, // green
  0xee8833, // orange
  0x9966cc, // purple
  0x33bbaa, // teal
  0xdd6699, // pink
  0xcccc33, // yellow
  0x66ccee, // sky blue
  0xaa4444, // dark red
  0x44aa88, // sea green
  0xcc7744, // rust
  0x8888dd, // lavender
  0x55bb77, // mint
  0xdd9944, // amber
  0xbb55aa, // magenta
  0x77aacc, // steel blue
  0xaaaa55, // olive
];

export function playerColor(joinSeq: number): number {
  return PALETTE[joinSeq % PALETTE.length];
}

export function playerColorCSS(joinSeq: number): string {
  const c = playerColor(joinSeq);
  return `#${c.toString(16).padStart(6, "0")}`;
}
