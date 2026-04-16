package board

import "errors"

// Sentinel errors returned by PlaceStone.
var (
	ErrOutOfBounds = errors.New("board: point is outside the grid")
	ErrOccupied    = errors.New("board: cell is already occupied")
	ErrSelfCapture = errors.New("board: move would self-capture")
	ErrKo          = errors.New("board: move repeats a previous board position (ko)")
)

// Color is the occupant of a cell.
// The zero value is Empty, so a freshly allocated Grid is valid without
// further initialisation.
type Color uint8

const (
	Empty Color = iota
	Black
	White
)

// Point is a board coordinate pair.
type Point struct{ X, Y int }

// Grid is an immutable snapshot of the board. Copy it freely; all
// mutations return a new Grid and leave the original unchanged.
type Grid struct {
	W, H  int
	cells []Color // flat row-major: index = Y*W + X
}

// New returns an empty W×H grid.
func New(w, h int) Grid {
	return Grid{W: w, H: h, cells: make([]Color, w*h)}
}

// InBounds reports whether (x, y) lies within the grid.
func (g Grid) InBounds(x, y int) bool {
	return x >= 0 && x < g.W && y >= 0 && y < g.H
}

// At returns the Color at (x, y). Panics if out of bounds.
func (g Grid) At(x, y int) Color {
	return g.cells[y*g.W+x]
}

// Equal reports whether g and other represent the same board position.
func (g Grid) Equal(other Grid) bool {
	if g.W != other.W || g.H != other.H {
		return false
	}
	for i, c := range g.cells {
		if c != other.cells[i] {
			return false
		}
	}
	return true
}

// set returns a new Grid with (x, y) changed to c.
func (g Grid) set(x, y int, c Color) Grid {
	next := Grid{W: g.W, H: g.H, cells: make([]Color, len(g.cells))}
	copy(next.cells, g.cells)
	next.cells[y*g.W+x] = c
	return next
}

// removeGroup returns a new Grid with every point in grp set to Empty.
func (g Grid) removeGroup(grp []Point) Grid {
	next := Grid{W: g.W, H: g.H, cells: make([]Color, len(g.cells))}
	copy(next.cells, g.cells)
	for _, p := range grp {
		next.cells[p.Y*next.W+p.X] = Empty
	}
	return next
}

// neighbors returns the up-to-four orthogonally adjacent points that lie
// within the grid bounds.
func (g Grid) neighbors(x, y int) []Point {
	pts := make([]Point, 0, 4)
	for _, d := range [4][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
		nx, ny := x+d[0], y+d[1]
		if g.InBounds(nx, ny) {
			pts = append(pts, Point{nx, ny})
		}
	}
	return pts
}

// group returns every point in the connected same-color group containing
// (x, y), using an iterative depth-first flood-fill.
// Returns nil when (x, y) is Empty.
func (g Grid) group(x, y int) []Point {
	color := g.At(x, y)
	if color == Empty {
		return nil
	}
	visited := make(map[Point]bool)
	stack := []Point{{x, y}}
	var result []Point
	for len(stack) > 0 {
		p := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if visited[p] {
			continue
		}
		visited[p] = true
		result = append(result, p)
		for _, nb := range g.neighbors(p.X, p.Y) {
			if !visited[nb] && g.At(nb.X, nb.Y) == color {
				stack = append(stack, nb)
			}
		}
	}
	return result
}

// liberties counts the distinct empty cells orthogonally adjacent to grp.
func (g Grid) liberties(grp []Point) int {
	seen := make(map[Point]bool)
	n := 0
	for _, p := range grp {
		for _, nb := range g.neighbors(p.X, p.Y) {
			if !seen[nb] && g.At(nb.X, nb.Y) == Empty {
				seen[nb] = true
				n++
			}
		}
	}
	return n
}

// opponent returns the opposing color in a two-player (Black/White) game.
func opponent(c Color) Color {
	if c == Black {
		return White
	}
	return Black
}

// PlaceStone attempts to place a stone of color c at (x, y) on grid g,
// applying standard Go rules: captures, self-capture rejection, and ko.
//
// The ko parameter is the board state the resulting position must not
// match; pass a zero-value Grid to skip the ko check entirely.
// Callers in package game are responsible for supplying the correct
// forbidden state — typically the board from two half-moves ago.
//
// On success it returns the updated Grid, the list of captured Points,
// and a nil error. The original grid is never modified.
func PlaceStone(g Grid, x, y int, c Color, ko Grid) (Grid, []Point, error) {
	if !g.InBounds(x, y) {
		return g, nil, ErrOutOfBounds
	}
	if g.At(x, y) != Empty {
		return g, nil, ErrOccupied
	}

	// Place the stone tentatively.
	next := g.set(x, y, c)

	// Capture any opponent groups that are now at zero liberties.
	// Only groups adjacent to the placed stone can be newly starved.
	// We re-check the cell color on each iteration because an earlier
	// capture may have already cleared a neighbor that shared a group
	// with the current candidate.
	opp := opponent(c)
	var captured []Point
	for _, nb := range next.neighbors(x, y) {
		if next.At(nb.X, nb.Y) != opp {
			continue
		}
		grp := next.group(nb.X, nb.Y)
		if next.liberties(grp) == 0 {
			captured = append(captured, grp...)
			next = next.removeGroup(grp)
		}
	}

	// Self-capture: the placed stone's own group must have at least one
	// liberty after all opponent captures are resolved.
	if next.liberties(next.group(x, y)) == 0 {
		return g, nil, ErrSelfCapture
	}

	// Ko: reject if the resulting position equals the forbidden state.
	if ko.W != 0 && next.Equal(ko) {
		return g, nil, ErrKo
	}

	return next, captured, nil
}
