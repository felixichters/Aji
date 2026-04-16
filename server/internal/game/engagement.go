package game

import "github.com/felixichters/Aji/server/internal/player"

// EngagementGraph is an undirected graph whose edges represent engaged
// pairs of players. Nodes appear only once they take part in an edge —
// isolated (unengaged) players are not tracked here.
type EngagementGraph struct {
	neighbors map[player.ID]map[player.ID]struct{}
}

// NewEngagementGraph returns an empty graph.
func NewEngagementGraph() *EngagementGraph {
	return &EngagementGraph{neighbors: make(map[player.ID]map[player.ID]struct{})}
}

// AddEdge inserts the undirected edge (a, b). It is idempotent and
// rejects self-loops. Returns true if a new edge was actually added.
func (g *EngagementGraph) AddEdge(a, b player.ID) bool {
	if a == b {
		return false
	}
	if _, ok := g.neighbors[a][b]; ok {
		return false
	}
	if g.neighbors[a] == nil {
		g.neighbors[a] = make(map[player.ID]struct{})
	}
	if g.neighbors[b] == nil {
		g.neighbors[b] = make(map[player.ID]struct{})
	}
	g.neighbors[a][b] = struct{}{}
	g.neighbors[b][a] = struct{}{}
	return true
}

// Has reports whether (a, b) is an edge.
func (g *EngagementGraph) Has(a, b player.ID) bool {
	_, ok := g.neighbors[a][b]
	return ok
}

// Neighbors returns the set of players engaged with p. The returned map
// is the graph's internal storage; callers must not mutate it.
func (g *EngagementGraph) Neighbors(p player.ID) map[player.ID]struct{} {
	return g.neighbors[p]
}

// Nodes returns every player currently present in the graph.
func (g *EngagementGraph) Nodes() []player.ID {
	out := make([]player.ID, 0, len(g.neighbors))
	for id := range g.neighbors {
		out = append(out, id)
	}
	return out
}
