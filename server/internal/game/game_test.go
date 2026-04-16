package game

import (
	"errors"
	"testing"

	"github.com/felixichters/Aji/server/internal/board"
	"github.com/felixichters/Aji/server/internal/player"
)

const testRadius = 5

func newTestGame(t *testing.T) *Game {
	t.Helper()
	return New(board.New(100, 100), testRadius)
}

func cell(x, y int) board.Cell { return board.Cell{X: x, Y: y} }

func mustJoin(t *testing.T, g *Game, id player.ID) *player.Player {
	t.Helper()
	p, err := g.Join(id)
	if err != nil {
		t.Fatalf("Join(%q): %v", id, err)
	}
	return p
}

func mustMove(t *testing.T, g *Game, id player.ID, c board.Cell) {
	t.Helper()
	if err := g.ApplyMove(id, c); err != nil {
		t.Fatalf("ApplyMove(%q, %v): %v", id, c, err)
	}
}

func expectErr(t *testing.T, g *Game, id player.ID, c board.Cell, want error) {
	t.Helper()
	err := g.ApplyMove(id, c)
	if !errors.Is(err, want) {
		t.Fatalf("ApplyMove(%q, %v) err = %v, want %v", id, c, err, want)
	}
}

// (1) The very first player in an empty world may place their opening
// stone anywhere — the bootstrap exemption.
func TestFirstPlayerFreeOpening(t *testing.T) {
	g := newTestGame(t)
	mustJoin(t, g, "A")
	mustMove(t, g, "A", cell(50, 50))
	if len(g.Players["A"].Stones) != 1 {
		t.Fatalf("expected 1 stone, got %d", len(g.Players["A"].Stones))
	}
	if len(g.Rotations) != 0 {
		t.Fatalf("expected no cliques yet, got %d", len(g.Rotations))
	}
}

// (2) After the opening stone a lone player is frozen until someone
// engages them.
func TestFirstPlayerBlockedUntilEngaged(t *testing.T) {
	g := newTestGame(t)
	mustJoin(t, g, "A")
	mustMove(t, g, "A", cell(50, 50))
	expectErr(t, g, "A", cell(51, 50), ErrNotEngaged)
}

// (3) A newcomer must place their first stone inside some existing
// player's region.
func TestSecondPlayerMustEngageOnFirstMove(t *testing.T) {
	g := newTestGame(t)
	mustJoin(t, g, "A")
	mustJoin(t, g, "B")
	mustMove(t, g, "A", cell(50, 50))

	expectErr(t, g, "B", cell(80, 80), ErrBootstrapMustEngage)
	mustMove(t, g, "B", cell(52, 50)) // inside A's region

	if !g.Graph.Has("A", "B") {
		t.Fatal("expected A-B engagement")
	}
	if len(g.Rotations) != 1 {
		t.Fatalf("expected 1 clique, got %d", len(g.Rotations))
	}
}

// (4) Two engaged players alternate strictly; same-player double moves
// are rejected.
func TestPairAlternation(t *testing.T) {
	g := newTestGame(t)
	mustJoin(t, g, "A")
	mustJoin(t, g, "B")
	mustMove(t, g, "A", cell(50, 50))
	mustMove(t, g, "B", cell(52, 50)) // engages A; {A,B}.next = A

	expectErr(t, g, "B", cell(53, 50), ErrNotYourTurn)
	mustMove(t, g, "A", cell(51, 50))
	expectErr(t, g, "A", cell(49, 50), ErrNotYourTurn)
	mustMove(t, g, "B", cell(53, 50))
}

// (5) Once engaged, a player may only place inside their own or an
// engaged neighbour's region.
func TestPlacementRestrictedToOwnOrEngagedRegion(t *testing.T) {
	g := newTestGame(t)
	mustJoin(t, g, "A")
	mustJoin(t, g, "B")
	mustMove(t, g, "A", cell(50, 50))
	mustMove(t, g, "B", cell(52, 50))

	// (90, 90) lies outside both regions.
	expectErr(t, g, "A", cell(90, 90), ErrOutsideRegion)
	// (56, 50) is only in B's region (dist 4 from B's (52,50); dist 6
	// from A's (50,50)). A is engaged with B, so this is legal.
	mustMove(t, g, "A", cell(56, 50))
}

// (6) In a path engagement A-B-C (no A-C edge), A and C are
// independent: either may move without waiting on the other; only B is
// coupled.
func TestPathIndependence(t *testing.T) {
	g := newTestGame(t)
	mustJoin(t, g, "A")
	mustJoin(t, g, "B")
	mustJoin(t, g, "C")

	// A and B engage, then walk both rotations so the set-up is clear.
	mustMove(t, g, "A", cell(10, 10))
	mustMove(t, g, "B", cell(14, 10)) // engages A; {A,B}.next = A
	mustMove(t, g, "A", cell(11, 10)) // advances -> next = B
	mustMove(t, g, "B", cell(15, 10)) // advances -> next = A
	mustMove(t, g, "A", cell(12, 10)) // advances -> next = B
	mustMove(t, g, "B", cell(16, 10)) // advances -> next = A

	// C's first stone engages B but not A. (19,10): distance 3 from
	// B's (16,10) (in range) and 7 from A's (12,10) (out of range).
	mustMove(t, g, "C", cell(19, 10))

	if _, ok := g.Rotations["A|B"]; !ok {
		t.Fatalf("expected {A,B} clique, got %v", keysOf(g.Rotations))
	}
	if _, ok := g.Rotations["B|C"]; !ok {
		t.Fatalf("expected {B,C} clique, got %v", keysOf(g.Rotations))
	}
	if _, ok := g.Rotations["A|B|C"]; ok {
		t.Fatalf("did not expect a triangle clique at this point")
	}

	// {A,B}.next = A, {B,C}.next = B.
	// Neither B nor C may move. A can.
	expectErr(t, g, "B", cell(17, 10), ErrNotYourTurn)
	expectErr(t, g, "C", cell(20, 10), ErrNotYourTurn)

	// A moves somewhere inside A's region that is NOT inside C's region
	// (so no accidental A-C engagement forms). (9,10) is dist 1 from
	// A's (10,10) and dist 10 from C's (19,10).
	mustMove(t, g, "A", cell(9, 10))
	// {A,B}.next = B. Now B may move (also B's turn in {B,C}).
	mustMove(t, g, "B", cell(17, 10))
	// After B, {A,B}.next = A and {B,C}.next = C. A and C are both
	// eligible; neither depends on the other.
	expectErr(t, g, "B", cell(18, 11), ErrNotYourTurn)
	mustMove(t, g, "C", cell(20, 10))
	// A's rotation was untouched by C's move: A can still move.
	mustMove(t, g, "A", cell(9, 11))

	// Finally sanity-check: no A-C edge formed at any point.
	if g.Graph.Has("A", "C") {
		t.Fatal("A-C should not be engaged in a path scenario")
	}
}

// (7) When A places a stone inside C's region while A-B and B-C are
// already engaged, the engagement graph becomes a triangle and the two
// pairwise cliques collapse into one 3-clique with a fresh cyclic
// rotation based on join order.
func TestTriangleFormation(t *testing.T) {
	g := newTestGame(t)
	mustJoin(t, g, "A")
	mustJoin(t, g, "B")
	mustJoin(t, g, "C")

	mustMove(t, g, "A", cell(10, 10))
	mustMove(t, g, "B", cell(13, 10)) // engages A; {A,B}.next = A
	mustMove(t, g, "A", cell(12, 10)) // next = B
	mustMove(t, g, "B", cell(14, 10)) // next = A
	mustMove(t, g, "C", cell(18, 10)) // engages B, not A; {B,C}.next = B

	if _, ok := g.Rotations["A|B|C"]; ok {
		t.Fatalf("did not expect triangle yet")
	}

	// It is A's turn. (17,10) lies in A's region (dist 5 from A's
	// (12,10)) and inside C's region (dist 1 from C's (18,10)), which
	// creates the new A-C edge and completes the triangle.
	mustMove(t, g, "A", cell(17, 10))

	if !g.Graph.Has("A", "C") {
		t.Fatal("expected A-C engagement after A plays in C's region")
	}
	if _, ok := g.Rotations["A|B|C"]; !ok {
		t.Fatalf("expected triangle clique, got %v", keysOf(g.Rotations))
	}
	if _, ok := g.Rotations["A|B"]; ok {
		t.Fatal("expected {A,B} to be subsumed by the triangle")
	}
	if _, ok := g.Rotations["B|C"]; ok {
		t.Fatal("expected {B,C} to be subsumed by the triangle")
	}

	rot := g.Rotations["A|B|C"]
	if !(len(rot.Order) == 3 && rot.Order[0] == "A" && rot.Order[1] == "B" && rot.Order[2] == "C") {
		t.Fatalf("unexpected rotation order: %v", rot.Order)
	}
	if rot.ToMove() != "B" {
		t.Fatalf("expected B to move next after A triggered the triangle, got %q", rot.ToMove())
	}

	expectErr(t, g, "A", cell(11, 11), ErrNotYourTurn)
	expectErr(t, g, "C", cell(20, 10), ErrNotYourTurn)
	mustMove(t, g, "B", cell(15, 10))
	if rot.ToMove() != "C" {
		t.Fatalf("expected C to move next after B, got %q", rot.ToMove())
	}
}

// (8) Moves that do not land inside any other player's region leave
// the engagement graph and the clique set unchanged.
func TestEngagementIsMonotonic(t *testing.T) {
	g := newTestGame(t)
	mustJoin(t, g, "A")
	mustJoin(t, g, "B")
	mustMove(t, g, "A", cell(10, 10))
	mustMove(t, g, "B", cell(13, 10))

	before := snapshotKeys(g.Rotations)
	// A plays inside own region but outside B's region: (7,10) is dist
	// 3 from A's (10,10) and dist 6 from B's (13,10).
	mustMove(t, g, "A", cell(7, 10))
	after := snapshotKeys(g.Rotations)

	if !sameSet(before, after) {
		t.Fatalf("clique set changed unexpectedly: before=%v after=%v", before, after)
	}
	if !g.Graph.Has("A", "B") || len(g.Graph.Nodes()) != 2 {
		t.Fatalf("unexpected engagement graph state")
	}
}

// (9) Region membership is the union of discs — two far-apart stones
// cover two disjoint areas, and cells between them are not in region.
func TestRegionIsUnionOfDiscs(t *testing.T) {
	stones := []board.Cell{cell(10, 10), cell(50, 50)}
	cases := []struct {
		c    board.Cell
		want bool
	}{
		{cell(10, 10), true},  // exactly on first stone
		{cell(12, 10), true},  // inside first disc
		{cell(50, 52), true},  // inside second disc
		{cell(15, 10), true},  // dist 5 from first — boundary (inclusive)
		{cell(16, 10), false}, // dist 6 from first — outside
		{cell(30, 30), false}, // between the two, far from both
	}
	for _, tc := range cases {
		got := InRegion(stones, tc.c, testRadius)
		if got != tc.want {
			t.Errorf("InRegion(%v) = %v, want %v", tc.c, got, tc.want)
		}
	}
}

// (10) Basic error paths: unknown player, out-of-bounds, occupied cell,
// duplicate join.
func TestBasicErrorPaths(t *testing.T) {
	g := newTestGame(t)
	if _, err := g.Join("A"); err != nil {
		t.Fatal(err)
	}
	if _, err := g.Join("A"); !errors.Is(err, ErrDuplicatePlayer) {
		t.Fatalf("duplicate join err = %v, want ErrDuplicatePlayer", err)
	}
	expectErr(t, g, "ghost", cell(1, 1), ErrUnknownPlayer)
	expectErr(t, g, "A", cell(-1, 0), ErrOutOfBounds)
	mustMove(t, g, "A", cell(50, 50))

	mustJoin(t, g, "B")
	expectErr(t, g, "B", cell(50, 50), ErrOccupied)
}

// helpers
func keysOf(m map[CliqueKey]*Rotation) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, string(k))
	}
	return out
}

func snapshotKeys(m map[CliqueKey]*Rotation) map[string]struct{} {
	out := make(map[string]struct{}, len(m))
	for k := range m {
		out[string(k)] = struct{}{}
	}
	return out
}

func sameSet(a, b map[string]struct{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if _, ok := b[k]; !ok {
			return false
		}
	}
	return true
}
