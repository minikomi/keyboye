package main

import (
	"github.com/minikomi/keyboye/internal/note"
	driver "github.com/minikomi/rtmididrv"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"os"
)

var winTitle string = "🎹"
var winWidth, winHeight int32 = 620, 60

type KeyboyeState struct {
	Octave      note.NoteModifier
	ActiveNotes map[sdl.Keycode]note.AbsoluteNote
}

var state = KeyboyeState{
	5,
	map[sdl.Keycode]note.AbsoluteNote{},
}

func getKeyboardColor(i int32) (r, g, b uint8) {
	switch i % 4 {
	case 0:
		return 255, 255, 210
	case 1:
		return 210, 255, 255
	case 2:
		return 255, 210, 210
	case 3:
		return 210, 255, 222
	}
	return 0, 0, 0
}

var blackOffsets = map[note.NoteModifier]int32{
	1:  8,
	3:  18,
	6:  38,
	8:  48,
	10: 58,
}

var whiteOffsets = map[note.NoteModifier]int32{
	0:  2,
	2:  12,
	4:  22,
	5:  32,
	7:  42,
	9:  52,
	11: 62,
}

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func run() int {

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	var window *sdl.Window
	var renderer *sdl.Renderer
	var font *ttf.Font

	// initialize window
	window, err := sdl.CreateWindow(winTitle, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		winWidth, winHeight, sdl.WINDOW_SHOWN)
	must(err)
	defer window.Destroy()

	// initialize renderer
	renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	must(err)
	defer renderer.Destroy()

	// initialize font, messages
	err = ttf.Init()
	must(err)
	font, err = ttf.OpenFont("./assets/tiny.ttf", 12)
	must(err)
	defer font.Close()

	// initialize midi
	drv, err := driver.New()
	must(err)
	defer drv.Close()

	ins, err := drv.Ins()
	must(err)

	outs, err := drv.Outs()
	must(err)

	Draw(renderer, font)

	if len(os.Args) == 2 && os.Args[1] == "list" {
		PrintInPorts(ins)
		PrintOutPorts(outs)
		return 0
	}

	out := outs[0]
	out.Open()
	wr := CreateMidiWriterTo(out)

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch ev := event.(type) {
			case *sdl.KeyboardEvent:
				HandleKeyEvent(ev, wr)
				Draw(renderer, font)
			case *sdl.QuitEvent:
				println("Quit")
				running = false
			}
		}
	}
	return 0
}

func main() {
	os.Exit(run())
}
