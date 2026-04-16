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
