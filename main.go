package main

import (
	"fmt"
	"github.com/gdamore/tcell"
	"os"
)

type state struct {
	x, y int
}

func intmin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func intmax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func clamp(v, l, u int) int {
	return intmax(l, intmin(v, u))
}

func move(sc tcell.Screen, st state, dx int, dy int) state {
	w, h := sc.Size()
	switch {
	case st.x == 0 && dx == -1:
		st.x = w - 1
		move(sc, st, 0, -1)
	case st.x == w-1 && dx == 1:
		st.x = 0
		move(sc, st, 0, 1)
	default:
		st.x = clamp(st.x+dx, 0, w)
		st.y = clamp(st.y+dy, 1, h)
	}
	return st
}

func initializeScreen() (sc tcell.Screen) {

	sc, e := tcell.NewScreen()

	if e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}

	if e = sc.Init(); e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}

	sc.Clear()
	sc.Sync()

	return sc
}

func drawStateDebug(sc tcell.Screen, st state) {
	guitxt := fmt.Sprintf("%d x %d", st.x, st.y)
	w, _ := sc.Size()
	for i := 0; i < w; i++ {
		sc.SetContent(i, 0, ' ', nil, 0)
	}
	for i, r := range guitxt {
		sc.SetContent(i, 0, r, nil, 0)
	}
	sc.Sync()
}

func drawKey(sc tcell.Screen, st state, ev *tcell.EventKey) {
	sc.SetContent(st.x, st.y, ev.Rune(), nil, 0)
	sc.Sync()
}

func main() {
	sc := initializeScreen()
	quit := make(chan struct{})

	st := state{0, 1}

	go func() {
		for {
			sc.ShowCursor(st.x, st.y)
			drawStateDebug(sc, st)
			ev := sc.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyEscape, tcell.KeyEnter:
					close(quit)
					return
				case tcell.KeyUp:
					st = move(sc, st, 0, -1)
				case tcell.KeyDown:
					st = move(sc, st, 0, 1)
				case tcell.KeyLeft:
					st = move(sc, st, -1, 0)
				case tcell.KeyRight:
					st = move(sc, st, 1, 0)
				default:
					drawKey(sc, st, ev)
					st = move(sc, st, 1, 0)
				}
			case *tcell.EventResize:
				sc.Sync()
			}
		}
	}()

	<-quit
	sc.Fini()
}
