package game

import "github.com/felixichters/Aji/server/internal/player"

// CliqueKey is the canonical identity of a clique; see Clique.Key.
type CliqueKey string

// Rotation is the turn pointer for one maximal clique. Order lists the
// clique's members by JoinSeq so the rotation is deterministic regardless
// of when the clique first formed. NextIdx is the position of the player
// whose turn it currently is.
type Rotation struct {
	Order   []player.ID
	NextIdx int
}

// ToMove returns the player currently expected to move.
func (r *Rotation) ToMove() player.ID { return r.Order[r.NextIdx] }

// Advance steps the pointer one position, wrapping at the end.
func (r *Rotation) Advance() {
	r.NextIdx = (r.NextIdx + 1) % len(r.Order)
}
