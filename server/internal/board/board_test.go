package board

import (
	"errors"
	"testing"
)

// place is a test helper that places a stone and fails the test immediately
// on any unexpected error.
func place(t *testing.T, g Grid, x, y int, c Color) Grid {
	t.Helper()
	next, _, err := PlaceStone(g, x, y, c, Grid{})
	if err != nil {
		t.Fatalf("place(%d,%d,%v): unexpected error: %v", x, y, c, err)
	}
	return next
}

// --- Construction -----------------------------------------------------------

func TestNew(t *testing.T) {
	g := New(3, 3)
	if g.W != 3 || g.H != 3 {
		t.Fatalf("New(3,3): got W=%d H=%d", g.W, g.H)
	}
	for y := 0; y < g.H; y++ {
		for x := 0; x < g.W; x++ {
			if g.At(x, y) != Empty {
				t.Errorf("At(%d,%d) = %v, want Empty", x, y, g.At(x, y))
			}
		}
	}
}

// --- InBounds ---------------------------------------------------------------

func TestInBounds(t *testing.T) {
	g := New(5, 5)
	cases := []struct {
		x, y int
		want  bool
	}{
		{0, 0, true},
		{4, 4, true},
		{-1, 0, false},
		{0, -1, false},
		{5, 0, false},
		{0, 5, false},
	}
	for _, tc := range cases {
		if got := g.InBounds(tc.x, tc.y); got != tc.want {
			t.Errorf("InBounds(%d,%d) = %v, want %v", tc.x, tc.y, got, tc.want)
		}
	}
}

// --- Basic placement --------------------------------------------------------

func TestPlaceStone_basic(t *testing.T) {
	g := New(5, 5)
	g2, _, err := PlaceStone(g, 2, 2, Black, Grid{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g2.At(2, 2) != Black {
		t.Errorf("At(2,2) = %v, want Black", g2.At(2, 2))
	}
	// Immutability: original must be unchanged.
	if g.At(2, 2) != Empty {
		t.Error("original grid was mutated")
	}
}

func TestPlaceStone_outOfBounds(t *testing.T) {
	g := New(3, 3)
	_, _, err := PlaceStone(g, 5, 0, Black, Grid{})
	if !errors.Is(err, ErrOutOfBounds) {
		t.Errorf("expected ErrOutOfBounds, got %v", err)
	}
}

func TestPlaceStone_occupied(t *testing.T) {
	g := New(3, 3)
	g = place(t, g, 1, 1, Black)
	_, _, err := PlaceStone(g, 1, 1, White, Grid{})
	if !errors.Is(err, ErrOccupied) {
		t.Errorf("expected ErrOccupied, got %v", err)
	}
}

// --- Captures ---------------------------------------------------------------

// TestCaptureSingle: a single White stone surrounded by Black is captured.
//
//	col: 0 1 2
//	row 0: . B .
//	row 1: B W B
//	row 2: . . .   ← Black plays (1,2), capturing White at (1,1)
func TestCaptureSingle(t *testing.T) {
	g := New(3, 3)
	g = place(t, g, 1, 0, Black)
	g = place(t, g, 0, 1, Black)
	g = place(t, g, 2, 1, Black)
	g = place(t, g, 1, 1, White)

	next, captured, err := PlaceStone(g, 1, 2, Black, Grid{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(captured) != 1 || captured[0] != (Point{1, 1}) {
		t.Errorf("captured = %v, want [{1 1}]", captured)
	}
	if next.At(1, 1) != Empty {
		t.Errorf("captured stone not removed")
	}
}

// TestCaptureGroup: a two-stone White group is captured when its last
// liberty is filled.
//
//	col: 0 1 2 3
//	row 0: B B B B
//	row 1: B W W .   ← Black plays (3,1), capturing {(1,1),(2,1)}
//	row 2: B B B B
func TestCaptureGroup(t *testing.T) {
	g := New(4, 3)
	for x := 0; x < 4; x++ {
		g = place(t, g, x, 0, Black)
		g = place(t, g, x, 2, Black)
	}
	g = place(t, g, 0, 1, Black)
	g = place(t, g, 1, 1, White)
	g = place(t, g, 2, 1, White)

	next, captured, err := PlaceStone(g, 3, 1, Black, Grid{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(captured) != 2 {
		t.Errorf("expected 2 captures, got %d: %v", len(captured), captured)
	}
	if next.At(1, 1) != Empty || next.At(2, 1) != Empty {
		t.Error("captured stones not removed from board")
	}
}

// TestCaptureReturnsZeroOnNoCapture: a move with no captures returns an
// empty slice (len 0); pin this so callers can safely range over it.
func TestCaptureReturnsZeroOnNoCapture(t *testing.T) {
	g := New(3, 3)
	_, captured, err := PlaceStone(g, 1, 1, Black, Grid{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(captured) != 0 {
		t.Errorf("expected no captures, got %v", captured)
	}
}

// --- Self-capture -----------------------------------------------------------

// TestSelfCapture: placing into a fully surrounded position is illegal
// when no opponent stones are captured.
//
//	col: 0 1 2
//	row 0: . B .
//	row 1: B . B
//	row 2: . B .   ← White at (1,1) has no liberties, no captures → illegal
func TestSelfCapture(t *testing.T) {
	g := New(3, 3)
	g = place(t, g, 1, 0, Black)
	g = place(t, g, 0, 1, Black)
	g = place(t, g, 2, 1, Black)
	g = place(t, g, 1, 2, Black)

	_, _, err := PlaceStone(g, 1, 1, White, Grid{})
	if !errors.Is(err, ErrSelfCapture) {
		t.Errorf("expected ErrSelfCapture, got %v", err)
	}
}

// TestCaptureBeforeSelfCapture: a move that looks like self-capture is
// legal when it simultaneously captures all surrounding opponent stones
// (captures are resolved before the self-capture check).
//
//	col: 0 1 2
//	row 0: B W B     each White stone is isolated (no White neighbor)
//	row 1: W . W     and has no outside liberties (all surrounded by B/OOB)
//	row 2: B W B
//
// Black plays (1,1): each adjacent White group has 0 liberties → all 4
// captured. Black then has 4 liberties via the vacated cells. Legal.
func TestCaptureBeforeSelfCapture(t *testing.T) {
	g := New(3, 3)
	g = place(t, g, 0, 0, Black)
	g = place(t, g, 2, 0, Black)
	g = place(t, g, 0, 2, Black)
	g = place(t, g, 2, 2, Black)
	g = place(t, g, 1, 0, White)
	g = place(t, g, 0, 1, White)
	g = place(t, g, 2, 1, White)
	g = place(t, g, 1, 2, White)

	next, captured, err := PlaceStone(g, 1, 1, Black, Grid{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(captured) != 4 {
		t.Errorf("expected 4 captures, got %d: %v", len(captured), captured)
	}
	if next.At(1, 1) != Black {
		t.Errorf("Black stone missing after capture")
	}
}

// --- Ko ---------------------------------------------------------------------

// TestKo verifies that the ko rule (ErrKo) fires when the resulting board
// position equals the explicitly supplied forbidden state.
//
// Classic ko shape on a 4×3 board:
//
//	col: 0 1 2 3
//	row 0: . B W .
//	row 1: B W . W   ← White at (1,1) has one liberty at (2,1)
//	row 2: . B W .
//
// White stones at (2,0), (3,1), (2,2) surround (2,1) on three sides, so
// when Black plays (2,1) and captures White, a White recapture at (1,1)
// would capture Black at (2,1) (whose remaining neighbors are all White)
// and restore the board to State A — triggering ko.
func TestKo(t *testing.T) {
	g := New(4, 3)
	// Black stones surrounding White at (1,1) on three sides.
	g = place(t, g, 1, 0, Black)
	g = place(t, g, 0, 1, Black)
	g = place(t, g, 1, 2, Black)
	// White stones surrounding the ko point (2,1) on three sides.
	g = place(t, g, 2, 0, White)
	g = place(t, g, 3, 1, White)
	g = place(t, g, 2, 2, White)
	// White at (1,1): the stone about to be captured.
	g = place(t, g, 1, 1, White)

	stateA := g // board before Black's capturing move — the forbidden state

	// Black plays (2,1), filling White's last liberty and capturing it.
	afterBlack, captured, err := PlaceStone(g, 2, 1, Black, Grid{})
	if err != nil {
		t.Fatalf("Black capture move: %v", err)
	}
	if len(captured) != 1 || captured[0] != (Point{1, 1}) {
		t.Fatalf("expected White at (1,1) captured, got %v", captured)
	}

	// White now wants to recapture at (1,1). This would capture Black at
	// (2,1) (whose other three neighbors are all White) and restore stateA.
	// The ko rule must reject this.
	_, _, err = PlaceStone(afterBlack, 1, 1, White, stateA)
	if !errors.Is(err, ErrKo) {
		t.Errorf("expected ErrKo on recapture, got %v", err)
	}
}

// TestKoSkipped: passing a zero-value Grid as the ko parameter allows the
// recapture that would otherwise be forbidden.
func TestKoSkipped(t *testing.T) {
	g := New(4, 3)
	g = place(t, g, 1, 0, Black)
	g = place(t, g, 0, 1, Black)
	g = place(t, g, 1, 2, Black)
	g = place(t, g, 2, 0, White)
	g = place(t, g, 3, 1, White)
	g = place(t, g, 2, 2, White)
	g = place(t, g, 1, 1, White)

	afterBlack, _, _ := PlaceStone(g, 2, 1, Black, Grid{})

	// With ko=Grid{} the immediate recapture is permitted.
	_, captured, err := PlaceStone(afterBlack, 1, 1, White, Grid{})
	if err != nil {
		t.Fatalf("recapture without ko check: %v", err)
	}
	if len(captured) != 1 || captured[0] != (Point{2, 1}) {
		t.Errorf("expected Black at (2,1) recaptured, got %v", captured)
	}
}

// --- Grid equality ----------------------------------------------------------

func TestEqual(t *testing.T) {
	a := New(3, 3)
	b := New(3, 3)
	if !a.Equal(b) {
		t.Error("two empty 3×3 grids should be equal")
	}
	a2 := place(t, a, 0, 0, Black)
	if a2.Equal(b) {
		t.Error("modified grid should not equal empty grid")
	}
	if a.Equal(New(4, 4)) {
		t.Error("different-size grids should not be equal")
	}
	// Original immutability check.
	if !a.Equal(b) {
		t.Error("original grid was mutated by place")
	}
}

// --- Neighbor topology ------------------------------------------------------

func TestNeighborCount(t *testing.T) {
	g := New(5, 5)
	cases := []struct {
		x, y int
		want  int
		label string
	}{
		{0, 0, 2, "corner"},
		{2, 0, 3, "top edge"},
		{0, 2, 3, "left edge"},
		{2, 2, 4, "interior"},
		{4, 4, 2, "corner"},
	}
	for _, tc := range cases {
		got := len(g.neighbors(tc.x, tc.y))
		if got != tc.want {
			t.Errorf("%s neighbors(%d,%d) = %d, want %d",
				tc.label, tc.x, tc.y, got, tc.want)
		}
	}
}
