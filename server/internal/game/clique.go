package game

import (
	"sort"
	"strings"

	"github.com/felixichters/Aji/server/internal/player"
)

// Clique is a set of mutually-engaged players, sorted by player ID for a
// canonical representation. Rotation order is derived separately from
// each player's JoinSeq (see rotation.go).
type Clique []player.ID

// Key returns a stable string identity for the clique (sorted IDs joined
// by '|'). Two cliques with the same member set produce the same key.
func (c Clique) Key() CliqueKey {
	ids := make([]string, len(c))
	for i, id := range c {
		ids[i] = string(id)
	}
	sort.Strings(ids)
	return CliqueKey(strings.Join(ids, "|"))
}

// Contains reports whether p is a member.
func (c Clique) Contains(p player.ID) bool {
	for _, m := range c {
		if m == p {
			return true
		}
	}
	return false
}

// MaximalCliques returns every maximal clique of size >= 2 in g, using
// the Bron–Kerbosch algorithm with pivot selection. Singleton nodes are
// filtered out because they carry no turn constraints.
func MaximalCliques(g *EngagementGraph) []Clique {
	p := make(map[player.ID]struct{}, len(g.neighbors))
	for id := range g.neighbors {
		p[id] = struct{}{}
	}
	var out []Clique
	bronKerbosch(g, map[player.ID]struct{}{}, p, map[player.ID]struct{}{}, &out)
	// Drop singletons and canonicalise ordering.
	filtered := out[:0]
	for _, c := range out {
		if len(c) < 2 {
			continue
		}
		sort.Slice(c, func(i, j int) bool { return c[i] < c[j] })
		filtered = append(filtered, c)
	}
	return filtered
}

func bronKerbosch(g *EngagementGraph, r, p, x map[player.ID]struct{}, out *[]Clique) {
	if len(p) == 0 && len(x) == 0 {
		members := make(Clique, 0, len(r))
		for id := range r {
			members = append(members, id)
		}
		*out = append(*out, members)
		return
	}
	// Pivot: vertex from P ∪ X with the most neighbours in P. Minimises
	// the number of recursive branches.
	pivot, ok := pickPivot(g, p, x)
	// Iterate P \ N(pivot).
	candidates := make([]player.ID, 0, len(p))
	for v := range p {
		if ok {
			if _, adj := g.neighbors[pivot][v]; adj {
				continue
			}
		}
		candidates = append(candidates, v)
	}
	for _, v := range candidates {
		newR := copySet(r)
		newR[v] = struct{}{}
		newP := intersect(p, g.neighbors[v])
		newX := intersect(x, g.neighbors[v])
		bronKerbosch(g, newR, newP, newX, out)
		delete(p, v)
		x[v] = struct{}{}
	}
}

func pickPivot(g *EngagementGraph, p, x map[player.ID]struct{}) (player.ID, bool) {
	var best player.ID
	bestCount := -1
	consider := func(u player.ID) {
		count := 0
		for v := range p {
			if _, ok := g.neighbors[u][v]; ok {
				count++
			}
		}
		if count > bestCount {
			best = u
			bestCount = count
		}
	}
	for u := range p {
		consider(u)
	}
	for u := range x {
		consider(u)
	}
	return best, bestCount >= 0
}

func copySet(s map[player.ID]struct{}) map[player.ID]struct{} {
	out := make(map[player.ID]struct{}, len(s))
	for k := range s {
		out[k] = struct{}{}
	}
	return out
}

func intersect(s map[player.ID]struct{}, adj map[player.ID]struct{}) map[player.ID]struct{} {
	out := make(map[player.ID]struct{})
	for k := range s {
		if _, ok := adj[k]; ok {
			out[k] = struct{}{}
		}
	}
	return out
}
