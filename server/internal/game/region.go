package game

import "github.com/felixichters/Aji/server/internal/board"

// DefaultRadius is the starting disc radius used when callers do not
// configure a specific value. Tuning is a follow-up task.
const DefaultRadius = 8

// InRegion reports whether cell c lies within Euclidean distance r of any
// stone in stones. r is compared against squared distance so no floating
// point is involved. A zero-stone slice always returns false.
func InRegion(stones []board.Cell, c board.Cell, r int) bool {
	if r < 0 {
		return false
	}
	rSq := r * r
	for _, s := range stones {
		dx := s.X - c.X
		dy := s.Y - c.Y
		if dx*dx+dy*dy <= rSq {
			return true
		}
	}
	return false
}
