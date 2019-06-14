package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strconv"
	"time"

	"github.com/gonutz/mines"
	"github.com/gonutz/prototype/draw"
	"github.com/gonutz/shuffle"
)

func main() {
	const (
		windowW, windowH = 950, 530
		tileSize         = 30
	)
	const (
		mainMenu = iota
		bestTimeScreen
		playing
	)
	const (
		beginner = iota
		advanced
		pro
	)

	var (
		state                  = mainMenu
		game                   *mines.Game
		offsetX, offsetY       int
		first                  bool
		startTime              time.Time
		winDuration            time.Duration
		selectedMenuItem       int
		bestTimesSec           = [9]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1}
		lastMouseX, lastMouseY int
	)

	if data, err := ioutil.ReadFile(bestFilePath()); err == nil {
		binary.Read(bytes.NewReader(data), binary.LittleEndian, &bestTimesSec)
	}
	defer func() {
		var buf bytes.Buffer
		if binary.Write(&buf, binary.LittleEndian, &bestTimesSec) == nil {
			ioutil.WriteFile(bestFilePath(), buf.Bytes(), 0666)
		}
	}()

	for i := range bestTimesSec {
		if bestTimesSec[i] == -1 {
			bestTimesSec[i] = 99*60*60 + 59*60 + 59
		}
	}

	rand.Seed(time.Now().UnixNano())

	mode := beginner
	newGame := func() {
		w, h, mineCount := 8, 8, 10
		if mode == pro {
			w, h, mineCount = 30, 16, 99
		} else if mode == advanced {
			w, h, mineCount = 16, 16, 40
		}
		game = mines.NewGame(w, h)
		m := make([]bool, w*h)
		for i := 0; i < mineCount; i++ {
			m[i] = true
		}
		shuffle.Shuffle(boolSlice(m), rand.Int)
		game.SetMines(m)
		offsetX = (windowW - w*tileSize) / 2
		offsetY = (windowH - h*tileSize) / 2
		first = true
		startTime = time.Now()
		winDuration = 0
	}

	check(draw.RunWindow("Minesweeper", windowW, windowH, func(window draw.Window) {
		if state == mainMenu {
			if window.WasKeyPressed(draw.KeyEscape) {
				window.Close()
				return
			}

			items := []string{
				" Beginner ",
				" Advanced ",
				" Professional ",
				" Best Times ",
			}

			if window.WasKeyPressed(draw.KeyDown) {
				selectedMenuItem = (selectedMenuItem + 1) % len(items)
			}
			if window.WasKeyPressed(draw.KeyUp) {
				selectedMenuItem = (selectedMenuItem + len(items) - 1) % len(items)
			}

			enter := func() {
				if selectedMenuItem == 3 {
					state = bestTimeScreen
				} else {
					mode = selectedMenuItem
					newGame()
					state = playing
				}
			}

			if window.WasKeyPressed(draw.KeyEnter) || window.WasKeyPressed(draw.KeyNumEnter) {
				enter()
				return
			}

			mx, my := window.MousePosition()
			for i, s := range items {
				const scale = 2.5
				w, h := window.GetScaledTextSize(s, scale)
				x := (windowW - w) / 2
				y := (windowH-h*len(items))/2 + i*h

				if mx != lastMouseX || my != lastMouseY {
					if mx >= x && my >= y && mx < x+w && my < y+h {
						selectedMenuItem = i
					}
				}

				if i == selectedMenuItem {
					window.FillRect(x, y, w, h, draw.DarkRed)
				}
				window.DrawScaledText(s, x, y, scale, draw.White)

				for _, click := range window.Clicks() {
					if click.Button == draw.LeftButton {
						mx, my := click.X, click.Y
						if mx >= x && my >= y && mx < x+w && my < y+h {
							selectedMenuItem = i
							enter()
							return
						}
					}
				}
			}

			lastMouseX, lastMouseY = mx, my
		} else if state == bestTimeScreen {
			if window.WasKeyPressed(draw.KeyEscape) {
				state = mainMenu
				return
			}

			caption := "Best Times"
			w, h := window.GetScaledTextSize(caption, 3)
			window.DrawScaledText(caption, (windowW-w)/2, 20, 3, draw.White)
			y := 20 + h + 50
			for mode, name := range []string{
				"Beginner",
				"Advanced",
				"Professional",
			} {
				times := bestTimesSec[3*mode : 3*mode+3]
				w, h := window.GetScaledTextSize(name, 2)
				window.DrawScaledText(name, (windowW-w)/2, y, 2, draw.White)
				y += h + 10
				for i, secs := range times {
					d := time.Duration(secs) * time.Second
					s := strconv.Itoa(i+1) + ". " + formatDuration(d)
					w, h := window.GetTextSize(s)
					window.DrawText(s, (windowW-w)/2, y, draw.White)
					y += h
				}
				y += 30
			}
		} else if state == playing {
			if window.WasKeyPressed(draw.KeyEscape) {
				state = mainMenu
				return
			}

			if window.WasKeyPressed(draw.KeyF2) {
				newGame()
				return
			}

			if !game.Lost() && !game.Won() {
				for _, click := range window.Clicks() {
					if click.X < offsetX || click.Y < offsetY {
						continue
					}
					tx := (click.X - offsetX) / tileSize
					ty := (click.Y - offsetY) / tileSize
					if tx >= game.Width() || ty >= game.Height() {
						continue
					}

					if click.Button == draw.LeftButton {
						if window.IsMouseDown(draw.RightButton) {
							game.Explode(tx, ty)
						} else {
							if game.Field(tx, ty).State() != mines.MarkedMine {
								game.Open(tx, ty)
								if first {
									for game.Field(tx, ty).IsMine() || game.Field(tx, ty).MineCount() != 0 {
										newGame()
										game.Open(tx, ty)
									}
								}
								first = false
							}
						}
					} else if click.Button == draw.RightButton {
						if window.IsMouseDown(draw.LeftButton) {
							game.Explode(tx, ty)
						} else {
							game.MarkNext(tx, ty)
						}
					} else if click.Button == draw.MiddleButton {
						game.Explode(tx, ty)
					}
				}

				if game.Won() && winDuration == 0 {
					winDuration = time.Now().Sub(startTime)
					secs := int32(winDuration / time.Second)
					times := bestTimesSec[3*mode : 3*mode+3]
					if secs < times[0] {
						times[0], times[1], times[2] = secs, times[0], times[1]
					} else if secs < times[1] {
						times[1], times[2] = secs, times[1]
					} else if secs < times[2] {
						times[2] = secs
					}
				}
				if game.Lost() {
					for y := 0; y < game.Height(); y++ {
						for x := 0; x < game.Width(); x++ {
							game.Open(x, y)
						}
					}
				}
			}

			window.FillRect(0, 0, windowW, windowH, draw.DarkGray)
			window.DrawRect(
				offsetX-2,
				offsetY-2,
				game.Width()*tileSize+4,
				game.Height()*tileSize+4,
				draw.Black,
			)
			for y := 0; y < game.Height(); y++ {
				for x := 0; x < game.Width(); x++ {
					left := offsetX + x*tileSize
					top := offsetY + y*tileSize
					f := game.Field(x, y)
					if f.State() == mines.Open {
						window.FillRect(
							left+1,
							top+1,
							tileSize-2,
							tileSize-2,
							draw.LightGray,
						)
						if f.IsMine() {
							mineSize := tileSize / 3
							window.DrawLine(
								left+(tileSize-mineSize)/2,
								top+(tileSize-mineSize)/2,
								left+(tileSize-mineSize)/2+mineSize-1,
								top+(tileSize-mineSize)/2+mineSize-1,
								draw.Black,
							)
							window.DrawLine(
								left+(tileSize-mineSize)/2+mineSize-1,
								top+(tileSize-mineSize)/2,
								left+(tileSize-mineSize)/2,
								top+(tileSize-mineSize)/2+mineSize-1,
								draw.Black,
							)
							window.FillEllipse(
								left+(tileSize-mineSize)/2,
								top+(tileSize-mineSize)/2,
								mineSize,
								mineSize,
								draw.DarkRed,
							)
							window.DrawEllipse(
								left+(tileSize-mineSize)/2,
								top+(tileSize-mineSize)/2,
								mineSize,
								mineSize,
								draw.Black,
							)
						} else {
							n := f.MineCount()
							if n > 0 {
								color := []draw.Color{
									draw.Black,      // 0
									draw.Blue,       // 1
									draw.DarkGreen,  // 2
									draw.Red,        // 3
									draw.DarkBlue,   // 4
									draw.DarkPurple, // 5
									draw.Brown,      // 6
									draw.DarkYellow, // 7
									draw.Cyan,       // 8
								}
								text := strconv.Itoa(n)
								w, h := window.GetTextSize(text)
								window.DrawText(
									text,
									left+(tileSize-w)/2,
									top+(tileSize-h)/2,
									color[n],
								)
							}
						}
					} else {
						window.FillRect(
							left,
							top,
							tileSize,
							tileSize,
							draw.LightGray,
						)
						window.FillRect(
							left+2,
							top+2,
							tileSize-2,
							tileSize-2,
							draw.DarkGray,
						)
						window.FillRect(
							left+2,
							top+2,
							tileSize-4,
							tileSize-4,
							draw.Gray,
						)
						if f.State() == mines.MarkedMine {
							poleX := left + 2*tileSize/3
							poleTop := top + 7
							poleBottom := top + tileSize - 7
							window.DrawLine(poleX, poleTop, poleX, poleBottom, draw.Black)
							window.DrawLine(poleX-5, poleBottom, poleX+3, poleBottom, draw.Black)
							window.DrawLine(poleX-8, poleTop+2, poleX, poleTop, draw.DarkRed)
							window.DrawLine(poleX-8, poleTop+2, poleX, poleTop+4, draw.DarkRed)
							window.DrawLine(poleX, poleTop, poleX, poleTop+4, draw.DarkRed)
						}
						if f.State() == mines.MarkedCandidate {
							text := "?"
							w, h := window.GetTextSize(text)
							window.DrawText(
								text,
								left+(tileSize-w)/2,
								top+(tileSize-h)/2,
								draw.Black,
							)
						}
					}
				}
			}

			var caption string
			if game.Won() {
				caption = "Won in " + formatDuration(winDuration)
			} else if game.Lost() {
				caption = "You Lost!"
			} else if !first {
				caption = formatDuration(time.Now().Sub(startTime))
			}
			w, _ := window.GetTextSize(caption)
			window.DrawText(caption, (windowW-w)/2, 5, draw.White)
		}
	}))
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type boolSlice []bool

func (x boolSlice) Len() int      { return len(x) }
func (x boolSlice) Swap(i, j int) { x[i], x[j] = x[j], x[i] }

func formatDuration(d time.Duration) string {
	sec := d / time.Second
	min := sec / 60
	h := min / 60
	return fmt.Sprintf("%02d:%02d:%02d", h, min%60, sec%60)
}
