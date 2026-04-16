package board

import "errors"

// Cell is a single coordinate on the board. Origin (0,0) is the top-left.
type Cell struct {
	X, Y int
}

// Board is the fixed-size grid. Occupancy is stored sparsely because the
// v0 grid is large (~200x200) and only a fraction of cells ever hold a
// stone.
type Board struct {
	w, h     int
	occupied map[Cell]struct{}
}

// New constructs an empty board of the given dimensions.
func New(w, h int) *Board {
	return &Board{w: w, h: h, occupied: make(map[Cell]struct{})}
}

// Size returns the board's width and height.
func (b *Board) Size() (int, int) { return b.w, b.h }

// InBounds reports whether c lies inside the grid.
func (b *Board) InBounds(c Cell) bool {
	return c.X >= 0 && c.X < b.w && c.Y >= 0 && c.Y < b.h
}

// Empty reports whether c is unoccupied. Out-of-bounds cells are not empty.
func (b *Board) Empty(c Cell) bool {
	if !b.InBounds(c) {
		return false
	}
	_, taken := b.occupied[c]
	return !taken
}

// Errors returned by Place.
var (
	ErrOutOfBounds = errors.New("board: cell out of bounds")
	ErrOccupied    = errors.New("board: cell already occupied")
)

// Place marks c as occupied. It returns an error if c is out of bounds or
// already holds a stone. Ownership is not tracked here — that belongs to
// higher layers (see package game).
func (b *Board) Place(c Cell) error {
	if !b.InBounds(c) {
		return ErrOutOfBounds
	}
	if _, taken := b.occupied[c]; taken {
		return ErrOccupied
	}
	b.occupied[c] = struct{}{}
	return nil
}
