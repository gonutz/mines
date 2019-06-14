/*
package mines contains the logic for a Minesweeper game.

A field can have multiple states:
	Closed          The player has not uncovered the field.
	Open            The player has uncovered the field.
	MarkedMine      The player has marked the field as mine (implies Closed).
	MarkedCandidate The player has marked the field as possible mine (implies Closed).
*/
package mines

// MineField is the mine count for a field that contains a mine. It is a value
// that is not possible for non-mine fields.
const MineField = 9

const (
	Closed = iota
	Open
	MarkedMine
	MarkedCandidate
)

func NewGame(width, height int) *Game {
	return &Game{
		width:  width,
		height: height,
		mine:   make([]bool, width*height),
		counts: make([]int, width*height),
		states: make([]int, width*height),
	}
}

type Game struct {
	width, height int
	mine          []bool
	counts        []int
	states        []int
	won, lost     bool
}

func (g *Game) Width() int  { return g.width }
func (g *Game) Height() int { return g.height }

func (g *Game) SetMines(mine []bool) {
	copy(g.mine, mine)
	g.updateCounts()
}

func (g *Game) updateCounts() {
	for y := 0; y < g.height; y++ {
		for x := 0; x < g.width; x++ {
			i := x + y*g.width
			if g.mine[i] {
				g.counts[i] = MineField
			} else {
				for nx := x - 1; nx <= x+1; nx++ {
					for ny := y - 1; ny <= y+1; ny++ {
						if g.mineAt(nx, ny) {
							g.counts[i]++
						}
					}
				}
			}
		}
	}
}

func (g *Game) mineAt(x, y int) bool {
	if x < 0 || y < 0 || x >= g.width || y >= g.height {
		return false
	}
	return g.mine[x+y*g.width]
}

func (g *Game) MineCounts() []int {
	return g.counts
}

func (g *Game) Won() bool {
	return g.won
}

func (g *Game) Lost() bool {
	return g.lost
}

func (g *Game) Open(x, y int) {
	if g.mineAt(x, y) {
		g.lost = true
	}

	q := [][2]int{[2]int{x, y}}
	for len(q) > 0 {
		x, y := q[0][0], q[0][1]
		q = q[1:]
		i := x + y*g.width
		g.states[i] = Open
		if g.counts[i] == 0 {
			ns := [][2]int{
				{x - 1, y - 1},
				{x - 1, y},
				{x - 1, y + 1},
				{x, y - 1},
				{x, y + 1},
				{x + 1, y - 1},
				{x + 1, y},
				{x + 1, y + 1},
			}
			for _, n := range ns {
				x, y := n[0], n[1]
				if x >= 0 && y >= 0 && x < g.width && y < g.height {
					i := x + y*g.width
					if g.states[i] != Open {
						q = append(q, n)
					}
				}
			}
		}
	}

	allClosedFieldsMines := true
	for i := range g.mine {
		if g.states[i] != Open && !g.mine[i] {
			allClosedFieldsMines = false
			break
		}
	}
	g.won = !g.lost && allClosedFieldsMines
}

func (g *Game) MarkNext(x, y int) {
	i := x + y*g.width
	switch g.states[i] {
	case Closed:
		g.states[i] = MarkedMine
	case MarkedMine:
		g.states[i] = Closed // MarkedCandidate
	case MarkedCandidate:
		g.states[i] = Closed
	}
}

func (g *Game) Explode(x, y int) {
	i := x + y*g.width
	if g.states[i] != Open || g.mine[i] || g.counts[i] == 0 {
		return
	}

	ns := [][2]int{
		{x - 1, y - 1},
		{x - 1, y},
		{x - 1, y + 1},
		{x, y - 1},
		{x, y + 1},
		{x + 1, y - 1},
		{x + 1, y},
		{x + 1, y + 1},
	}
	markCount := 0
	for _, n := range ns {
		x, y := n[0], n[1]
		if x >= 0 && y >= 0 && x < g.width && y < g.height {
			i := x + y*g.width
			if g.states[i] == MarkedMine {
				markCount++
			}
		}
	}

	if markCount == g.counts[i] {
		for _, n := range ns {
			x, y := n[0], n[1]
			if x >= 0 && y >= 0 && x < g.width && y < g.height {
				i := x + y*g.width
				if g.states[i] != MarkedMine {
					g.Open(x, y)
				}
			}
		}
	}
}

func (g *Game) Field(x, y int) Field {
	return field{
		state:     g.states[x+y*g.width],
		isMine:    g.mine[x+y*g.width],
		mineCount: g.counts[x+y*g.width],
	}
}

type Field interface {
	State() int
	IsMine() bool
	MineCount() int
}

type field struct {
	state     int
	isMine    bool
	mineCount int
}

func (f field) State() int     { return f.state }
func (f field) IsMine() bool   { return f.isMine }
func (f field) MineCount() int { return f.mineCount }
