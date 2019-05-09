package main

import (
	"fmt"
	"github.com/gomidi/connect"
	"github.com/gomidi/midi"
	"github.com/gomidi/midi/midimessage/channel"
	//	. "github.com/gomidi/midi/midimessage/channel"
	"github.com/gomidi/midi/midiwriter"
	driver "github.com/minikomi/rtmididrv"
	"github.com/veandco/go-sdl2/sdl"
	"io"
	"os"
)

var winTitle string = "Go-SDL2 Render"
var winWidth, winHeight int32 = 800, 600

type state struct {
}

var keyMap = map[sdl.Keycode]uint8{
	sdl.K_a: 63,
	sdl.K_s: 64,
	sdl.K_d: 65,
	sdl.K_f: 66,
}

func handleKeyEvent(ev *sdl.KeyboardEvent, wr *Writer) {
	fmt.Printf("[%d ms] Keyboard\ttype:%d\tsym:%c\tmodifiers:%d\tstate:%d\trepeat:%d\n",
		ev.Timestamp, ev.Type, ev.Keysym.Sym, ev.Keysym.Mod, ev.State, ev.Repeat)
	note, ok := keyMap[ev.Keysym.Sym]

	if ok {
		fmt.Println(note)
		if ev.State == 1 {
			if ev.Repeat != 1 {
				wr.NoteOn(note, 90)
			}
		} else {
			wr.NoteOff(note)
		}
	} else {
		fmt.Printf("%d %d\n", sdl.GetScancodeFromKey(ev.Keysym.Sym), sdl.K_a)
	}
}

func draw(renderer *sdl.Renderer) {
	var points []sdl.Point
	var rect sdl.Rect
	var rects []sdl.Rect

	renderer.Clear()

	renderer.SetDrawColor(255, 255, 255, 255)
	renderer.DrawPoint(150, 300)

	renderer.SetDrawColor(0, 0, 255, 255)
	renderer.DrawLine(0, 0, 200, 200)

	points = []sdl.Point{{0, 0}, {100, 300}, {100, 300}, {200, 0}}
	renderer.SetDrawColor(255, 255, 0, 255)
	renderer.DrawLines(points)

	rect = sdl.Rect{300, 0, 200, 200}
	renderer.SetDrawColor(255, 0, 0, 255)
	renderer.DrawRect(&rect)

	rects = []sdl.Rect{{400, 400, 100, 100}, {550, 350, 200, 200}}
	renderer.SetDrawColor(0, 255, 255, 255)
	renderer.DrawRects(rects)

	rect = sdl.Rect{250, 250, 200, 200}
	renderer.SetDrawColor(0, 255, 0, 255)
	renderer.FillRect(&rect)

	rects = []sdl.Rect{{500, 300, 100, 100}, {200, 300, 200, 200}}
	renderer.SetDrawColor(255, 0, 255, 255)
	renderer.FillRects(rects)

	renderer.Present()

}

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}

type midiWriter struct {
	wr              midi.Writer
	ch              channel.Channel
	noteState       [16][128]bool
	noConsolidation bool
}

type Writer struct {
	*midiWriter
}

func NewWriter(dest io.Writer, options ...midiwriter.Option) *Writer {
	options = append(
		[]midiwriter.Option{
			midiwriter.NoRunningStatus(),
		}, options...)

	wr := midiwriter.New(dest, options...)
	return &Writer{&midiWriter{wr: wr, ch: channel.Channel0}}
}

type outWriter struct {
	out connect.Out
}

func (w *outWriter) Write(b []byte) (int, error) {
	return len(b), w.out.Send(b)
}

func writeTo(out connect.Out) *Writer {
	return NewWriter(&outWriter{out})
}

func (w *midiWriter) NoteOn(key, veloctiy uint8) error {
	return w.Write(w.ch.NoteOn(key, veloctiy))
}

func (w *midiWriter) NoteOff(key uint8) error {
	return w.Write(w.ch.NoteOff(key))
}

func (w *midiWriter) Write(msg midi.Message) error {
	if w.noConsolidation {
		return w.wr.Write(msg)
	}
	switch m := msg.(type) {
	case channel.NoteOn:
		if m.Velocity() > 0 && w.noteState[m.Channel()][m.Key()] {
			return fmt.Errorf("can't write %s. note already running.", msg)
		}
		if m.Velocity() == 0 && !w.noteState[m.Channel()][m.Key()] {
			return fmt.Errorf("can't write %s. note is not running.", msg)
		}
		w.noteState[m.Channel()][m.Key()] = m.Velocity() > 0
	case channel.NoteOff:
		if !w.noteState[m.Channel()][m.Key()] {
			return fmt.Errorf("can't write %s. note is not running.", msg)
		}
		w.noteState[m.Channel()][m.Key()] = false
	case channel.NoteOffVelocity:
		if !w.noteState[m.Channel()][m.Key()] {
			return fmt.Errorf("can't write %s. note is not running.", msg)
		}
		w.noteState[m.Channel()][m.Key()] = false
	}
	return w.wr.Write(msg)
}

func run() int {

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()
	var window *sdl.Window
	var renderer *sdl.Renderer
	window, err := sdl.CreateWindow(winTitle, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		winWidth, winHeight, sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create window: %s\n", err)
		return 1
	}
	defer window.Destroy()
	renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create renderer: %s\n", err)
		return 2
	}

	defer renderer.Destroy()

	draw(renderer)

	// midi
	drv, err := driver.New()
	must(err)
	defer drv.Close()

	ins, err := drv.Ins()
	must(err)

	outs, err := drv.Outs()
	must(err)

	if len(os.Args) == 2 && os.Args[1] == "list" {
		printInPorts(ins)
		printOutPorts(outs)
		return 0
	}

	out := outs[0]
	out.Open()
	wr := writeTo(out)

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch ev := event.(type) {
			case *sdl.KeyboardEvent:
				handleKeyEvent(ev, wr)
			case *sdl.QuitEvent:
				println("Quit")
				running = false
				break
			}
		}
	}
	return 0
}

func printPort(port connect.Port) {
	fmt.Printf("[%v] %s\n", port.Number(), port.String())
}

func printInPorts(ports []connect.In) {
	fmt.Printf("MIDI IN Ports\n")
	for _, port := range ports {
		printPort(port)
	}
	fmt.Printf("\n\n")
}

func printOutPorts(ports []connect.Out) {
	fmt.Printf("MIDI OUT Ports\n")
	for _, port := range ports {
		printPort(port)
	}
	fmt.Printf("\n\n")
}

func main() {
	os.Exit(run())
}
