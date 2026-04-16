package game

import (
	"errors"
	"sort"

	"github.com/felixichters/Aji/server/internal/board"
	"github.com/felixichters/Aji/server/internal/player"
)

// Sentinel errors returned by Join, ValidateMove, and ApplyMove.
var (
	ErrUnknownPlayer       = errors.New("game: unknown player")
	ErrDuplicatePlayer     = errors.New("game: player already joined")
	ErrOutOfBounds         = errors.New("game: cell out of bounds")
	ErrOccupied            = errors.New("game: cell already occupied")
	ErrNotYourTurn         = errors.New("game: not your turn in one of your local games")
	ErrNotEngaged          = errors.New("game: player has no engagement and cannot move")
	ErrOutsideRegion       = errors.New("game: cell is outside own region and any engaged region")
	ErrBootstrapMustEngage = errors.New("game: first stone must land inside another player's region")
)

// Game is the authoritative rule engine. It owns the engagement graph
// and the rotation state of every active maximal clique. It is not
// goroutine-safe on its own; callers (package world) serialise access.
type Game struct {
	Radius      int
	Board       *board.Board
	Players     map[player.ID]*player.Player
	JoinCounter int
	Graph       *EngagementGraph
	Rotations   map[CliqueKey]*Rotation
}

// New constructs an empty game backed by b and using radius r for the
// union-of-discs region test.
func New(b *board.Board, r int) *Game {
	return &Game{
		Radius:    r,
		Board:     b,
		Players:   make(map[player.ID]*player.Player),
		Graph:     NewEngagementGraph(),
		Rotations: make(map[CliqueKey]*Rotation),
	}
}

// Join registers id as a new player and returns the Player record. It
// errors if id is already present.
func (g *Game) Join(id player.ID) (*player.Player, error) {
	if _, dup := g.Players[id]; dup {
		return nil, ErrDuplicatePlayer
	}
	p := player.New(id, g.JoinCounter)
	g.JoinCounter++
	g.Players[id] = p
	return p, nil
}

// ValidateMove checks whether p may legally place a stone at c. No state
// is mutated. See ApplyMove for the full move pipeline.
func (g *Game) ValidateMove(p player.ID, c board.Cell) error {
	pl, ok := g.Players[p]
	if !ok {
		return ErrUnknownPlayer
	}
	if !g.Board.InBounds(c) {
		return ErrOutOfBounds
	}
	if !g.Board.Empty(c) {
		return ErrOccupied
	}

	// Bootstrap branch: the player has no stones yet.
	if len(pl.Stones) == 0 {
		if g.anyOtherPlayerHasStones(p) {
			if !g.cellHitsAnyOtherRegion(p, c) {
				return ErrBootstrapMustEngage
			}
		}
		// Either another player's region is touched (engagement will form on
		// apply) or this is the very first stone in the world (free opening).
		return nil
	}

	// Post-bootstrap branch: the player already has stones.
	cliques := g.cliquesContaining(p)
	if len(cliques) == 0 {
		// Has stones but no engagements = lone opener waiting to be engaged.
		return ErrNotEngaged
	}
	for key := range cliques {
		if g.Rotations[key].ToMove() != p {
			return ErrNotYourTurn
		}
	}
	if g.cellInOwnOrEngagedRegion(p, pl, c) {
		return nil
	}
	return ErrOutsideRegion
}

// ApplyMove validates and executes a move. On success the board, the
// player's stone list, the engagement graph, and the rotation map are
// all updated consistently.
func (g *Game) ApplyMove(p player.ID, c board.Cell) error {
	if err := g.ValidateMove(p, c); err != nil {
		return err
	}
	pl := g.Players[p]
	preCliques := g.cliquesContaining(p)

	if err := g.Board.Place(c); err != nil {
		// Shouldn't happen — validation covers the same conditions. Surface
		// the board's error if it ever does.
		return err
	}
	pl.Stones = append(pl.Stones, c)

	// Pick up any new engagements caused by the placement.
	newEdge := false
	for qid, q := range g.Players {
		if qid == p {
			continue
		}
		if len(q.Stones) == 0 {
			continue
		}
		if !InRegion(q.Stones, c, g.Radius) {
			continue
		}
		if g.Graph.AddEdge(p, qid) {
			newEdge = true
		}
	}

	if newEdge {
		g.recomputeCliques(p)
	}

	// Pre-existing cliques that still exist get their rotation advanced
	// because p has just moved in them. Cliques that vanished (subsumed
	// by a bigger clique) are already absent from g.Rotations; new
	// cliques had their NextIdx seeded past p by recomputeCliques and
	// must not be advanced again.
	for key := range preCliques {
		if rot, ok := g.Rotations[key]; ok {
			rot.Advance()
		}
	}
	return nil
}

// cliquesContaining returns the set of clique keys that include p.
func (g *Game) cliquesContaining(p player.ID) map[CliqueKey]struct{} {
	out := make(map[CliqueKey]struct{})
	for key, rot := range g.Rotations {
		for _, m := range rot.Order {
			if m == p {
				out[key] = struct{}{}
				break
			}
		}
	}
	return out
}

func (g *Game) anyOtherPlayerHasStones(self player.ID) bool {
	for id, q := range g.Players {
		if id == self {
			continue
		}
		if len(q.Stones) > 0 {
			return true
		}
	}
	return false
}

func (g *Game) cellHitsAnyOtherRegion(self player.ID, c board.Cell) bool {
	for id, q := range g.Players {
		if id == self {
			continue
		}
		if len(q.Stones) == 0 {
			continue
		}
		if InRegion(q.Stones, c, g.Radius) {
			return true
		}
	}
	return false
}

func (g *Game) cellInOwnOrEngagedRegion(self player.ID, pl *player.Player, c board.Cell) bool {
	if InRegion(pl.Stones, c, g.Radius) {
		return true
	}
	for qid := range g.Graph.Neighbors(self) {
		q, ok := g.Players[qid]
		if !ok {
			continue
		}
		if InRegion(q.Stones, c, g.Radius) {
			return true
		}
	}
	return false
}

// recomputeCliques re-enumerates maximal cliques after a new engagement
// edge has been added, then reconciles g.Rotations: vanished cliques are
// dropped, surviving cliques keep their rotation untouched, and
// brand-new cliques get a fresh Rotation whose NextIdx points at the
// member immediately after the mover (so the mover's just-executed move
// counts as their turn in that new clique).
func (g *Game) recomputeCliques(mover player.ID) {
	fresh := MaximalCliques(g.Graph)
	seen := make(map[CliqueKey]Clique, len(fresh))
	for _, cl := range fresh {
		seen[cl.Key()] = cl
	}
	for key := range g.Rotations {
		if _, ok := seen[key]; !ok {
			delete(g.Rotations, key)
		}
	}
	for key, cl := range seen {
		if _, ok := g.Rotations[key]; ok {
			continue
		}
		g.Rotations[key] = g.seedRotation(cl, mover)
	}
}

func (g *Game) seedRotation(cl Clique, mover player.ID) *Rotation {
	order := make([]player.ID, len(cl))
	copy(order, cl)
	sort.Slice(order, func(i, j int) bool {
		return g.Players[order[i]].JoinSeq < g.Players[order[j]].JoinSeq
	})
	next := 0
	for i, id := range order {
		if id == mover {
			next = (i + 1) % len(order)
			break
		}
	}
	return &Rotation{Order: order, NextIdx: next}
}
