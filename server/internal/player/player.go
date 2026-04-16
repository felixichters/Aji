package player

import "github.com/felixichters/Aji/server/internal/board"

// ID identifies a player. Callers supply it (UUID, short handle, whatever);
// this package does not mint IDs.
type ID string

// Player is a participant's in-world state. Stones is the list of cells
// the player has placed, in placement order. JoinSeq is a monotonic
// counter assigned by the game when the player joins; it defines the
// canonical rotation order inside any clique the player takes part in.
type Player struct {
	ID      ID
	Stones  []board.Cell
	JoinSeq int
}

// New returns a freshly joined player with the given identity and join
// sequence number and no stones placed.
func New(id ID, joinSeq int) *Player {
	return &Player{ID: id, JoinSeq: joinSeq}
}
