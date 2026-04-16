# Turn rules and engagement

Spatial turn gating. Rule engine [`server/internal/game`]
(../server/internal/game).

## Terms

- **Stone.** A placement on a cell, owned by a single player.
- **Radius `R`.** Configurable per world; `game.DefaultRadius` is the
  starting value.
- **Region** of player `P`. The union of closed Euclidean discs of
  radius `R` centred on every one of `P`'s stones. A cell `c` is in
  `P`'s region iff `||c − s|| ≤ R` for some stone `s` owned by `P`.
  (Distance is compared squared to avoid floats.)
- **Engagement.** A symmetric, monotonic relation between two players.
  The edge `(P, Q)` is added when `P` places a stone inside `Q`'s
  region. Once added, it is not removed. (Captures are not implemented
  in v0; they would be the first thing to break monotonicity.)
- **Local game.** A *maximal clique* in the engagement graph. Two
  engaged players form a 2-clique; three pairwise-engaged players form
  a 3-clique that subsumes its pair-cliques; and so on.

## Placement legality

A move by player `P` at cell `c` is legal iff **all** of the following
hold:

1. `c` is inside the board.
2. `c` is unoccupied.
3. Turn gating:
   - If `P` has zero stones: no turn gating (bootstrap — see below).
   - Otherwise: for every local game (max clique) `P` belongs to, `P`
     is the one player currently *to move* in that clique's rotation.
4. Spatial rule:
   - If `P` has stones and is engaged with at least one other player:
     `c` must lie in `P`'s region OR in the region of some player that
     `P` is currently engaged with.
   - If `P` has stones and is engaged with nobody: the move is
     illegal (`ErrNotEngaged`).
   - If `P` has zero stones and at least one other player has stones:
     `c` must be inside some other player's region (bootstrap —
     `ErrBootstrapMustEngage`).
   - If `P` has zero stones and no other player has any stones either:
     `c` may be anywhere on the board. This is the free opening
     exemption, available exactly once, to the very first stone ever
     played in the world.

## Rotation inside a local game

Each maximal clique carries its own cyclic rotation. Members are
ordered by `JoinSeq` — the monotonic counter assigned when each player
joined the world — so the rotation is deterministic and independent of
when the clique first formed.

When a new max clique `K` comes into existence after a move by player
`P`, its rotation is seeded so `P` is treated as having just moved:
`NextIdx` points at the member immediately after `P` in the `JoinSeq`
ordering. If an existing clique survives the move unchanged, its
rotation is preserved and advanced by one step.

When a new edge bridges cliques — e.g. the A-C edge that completes a
triangle A-B-C — the two smaller cliques are dropped because they are
no longer maximal; the single larger clique takes over with a fresh
rotation.

## Worked examples

**Solo opener.** Player A joins an empty world and places one stone
anywhere. A is now blocked: no further moves until someone engages
them.

**Pair.** Player B joins and places their first stone inside A's
region. The edge `A-B` forms, the clique `{A, B}` is seeded. Because
B just moved, A is next.

**Path A-B-C.** After A-B and B-C are engaged but A-C is not, there
are two max cliques: `{A, B}` and `{B, C}`. A and C are independent —
neither waits on the other. B is coupled: B moves only when B is
"to move" in both rotations.

**Triangle.** If A later places a stone inside C's region (reached
via B's region, which A is allowed to play into because A is engaged
with B), the edge A-C forms, `{A, B}` and `{B, C}` are subsumed by
`{A, B, C}`. The new 3-clique's rotation is seeded in `JoinSeq` order
starting at the player after A — the standard 3-way cyclic
alternation.

## Out of scope (v0)

- Captures, ko, suicide. These would make engagement non-monotonic —
  region overlaps could disappear when a stone is captured.
- Obstacles, shrinking zone, accounts.
- Concurrent move resolution beyond the single-process mutex in
  `internal/world`.
- Wire protocol and client rendering of regions / rotations / turn
  state — the rule engine is in-process only for now.