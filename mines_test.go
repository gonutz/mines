package mines_test

import (
	"strconv"
	"testing"

	"github.com/gonutz/check"
	"github.com/gonutz/mines"
)

func TestGame(t *testing.T) {
	const w, h = 8, 4
	g := mines.NewGame(w, h)
	check.Eq(t, g.Width(), w)
	check.Eq(t, g.Height(), h)
	const x, o = true, false
	g.SetMines([]bool{
		o, o, o, o, o, o, o, o,
		o, o, o, o, o, o, o, o,
		o, x, o, o, x, o, o, o,
		o, o, o, o, o, x, o, o,
	})
	check.Eq(t, g.MineCounts(), []int{
		0, 0, 0, 0, 0, 0, 0, 0,
		1, 1, 1, 1, 1, 1, 0, 0,
		1, 9, 1, 1, 9, 2, 1, 0,
		1, 1, 1, 1, 2, 9, 1, 0,
	})
	check.Eq(t, g.Won(), false)
	check.Eq(t, g.Lost(), false)

	check.Eq(t, gameToString(g), `
........
........
........
........
`)

	for y := 0; y < g.Height(); y++ {
		for x := 0; x < g.Width(); x++ {
			g.Open(x, y)
		}
	}
	check.Eq(t, gameToString(g), `
00000000
11111100
1x11x210
11112x10
`)
}

func gameToString(g *mines.Game) string {
	var s string

	s += "\n"
	for y := 0; y < g.Height(); y++ {
		for x := 0; x < g.Width(); x++ {
			f := g.Field(x, y)
			switch f.State() {
			case mines.Open:
				if f.IsMine() {
					s += "x"
				} else {
					s += strconv.Itoa(f.MineCount())
				}
			case mines.Closed:
				s += "."
			case mines.MarkedMine:
				s += "o"
			case mines.MarkedCandidate:
				s += "?"
			}
		}
		s += "\n"
	}

	return s
}
