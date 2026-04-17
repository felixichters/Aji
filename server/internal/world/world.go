package world

import (
	"sync"

	"github.com/felixichters/Aji/server/internal/board"
	"github.com/felixichters/Aji/server/internal/game"
	"github.com/felixichters/Aji/server/internal/player"
)

// World is the single in-memory runtime aggregate. It owns the board
// and the rule engine and serialises concurrent access through a mutex
// so the rule logic in package game can stay lock-free.
type World struct {
	mu    sync.Mutex
	board *board.Board
	game  *game.Game
}

// New constructs an empty world with the given board dimensions and
// region radius.
func New(width, height, radius int) *World {
	b := board.New(width, height)
	return &World{
		board: b,
		game:  game.New(b, radius),
	}
}

// Join registers a new player.
func (w *World) Join(id player.ID) (*player.Player, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.game.Join(id)
}

// PlaceStone applies a move on behalf of player id.
func (w *World) PlaceStone(id player.ID, c board.Cell) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.game.ApplyMove(id, c)
}

// BoardSize returns the board's width and height.
func (w *World) BoardSize() (int, int) {
	return w.board.Size()
}

// Radius returns the engagement region radius.
func (w *World) Radius() int {
	return w.game.Radius
}

// PlayerSnapshot is the domain-level representation of a player's state.
type PlayerSnapshot struct {
	ID      player.ID
	JoinSeq int
	Stones  []board.Cell
}

// CliqueSnapshot is the domain-level representation of a clique's turn state.
type CliqueSnapshot struct {
	Members []player.ID
	ToMove  player.ID
}

// EngagementSnapshot is a pair of engaged player IDs.
type EngagementSnapshot struct {
	A, B player.ID
}

// Snapshot captures the full game state. It holds the mutex for the
// duration of the copy.
type Snapshot struct {
	Players     []PlayerSnapshot
	Cliques     []CliqueSnapshot
	Engagements []EngagementSnapshot
}

// Snapshot returns the current game state as a domain-level snapshot.
func (w *World) Snapshot() Snapshot {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Players.
	players := make([]PlayerSnapshot, 0, len(w.game.Players))
	for _, p := range w.game.Players {
		stones := make([]board.Cell, len(p.Stones))
		copy(stones, p.Stones)
		players = append(players, PlayerSnapshot{
			ID:      p.ID,
			JoinSeq: p.JoinSeq,
			Stones:  stones,
		})
	}

	// Cliques (from rotation map).
	cliques := make([]CliqueSnapshot, 0, len(w.game.Rotations))
	for _, rot := range w.game.Rotations {
		members := make([]player.ID, len(rot.Order))
		copy(members, rot.Order)
		cliques = append(cliques, CliqueSnapshot{
			Members: members,
			ToMove:  rot.ToMove(),
		})
	}

	// Engagement edges (deduplicate by emitting only when a < b).
	var engagements []EngagementSnapshot
	for _, n := range w.game.Graph.Nodes() {
		for nb := range w.game.Graph.Neighbors(n) {
			if n < nb {
				engagements = append(engagements, EngagementSnapshot{A: n, B: nb})
			}
		}
	}

	return Snapshot{
		Players:     players,
		Cliques:     cliques,
		Engagements: engagements,
	}
}
