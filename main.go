package main

import (
	"fmt"
	"github.com/gomidi/connect"
	"github.com/gomidi/midi"
	"github.com/gomidi/midi/midimessage/channel"
	"github.com/gomidi/midi/midiwriter"
	"github.com/minikomi/keyboye/internal/note"
	driver "github.com/minikomi/rtmididrv"
	"github.com/veandco/go-sdl2/sdl"
	"io"
	"os"
)

var winTitle string = "🎹"
var winWidth, winHeight int32 = 800, 60

type KeyboyeState struct {
	Octave      note.NoteModifier
	ActiveNotes map[sdl.Keycode]note.AbsoluteNote
}

var state = KeyboyeState{
	5,
	map[sdl.Keycode]note.AbsoluteNote{},
}

func getAbsoluteNote(noteModifier note.NoteModifier) note.AbsoluteNote {
	return note.AbsoluteNote(noteModifier + state.Octave*12)
}

var keyToNote = map[sdl.Keycode]note.NoteModifier{
	sdl.K_a: note.C,
	sdl.K_w: note.CSharp,
	sdl.K_s: note.D,
	sdl.K_e: note.DSharp,
	sdl.K_d: note.E,
	sdl.K_f: note.F,
	sdl.K_t: note.FSharp,
	sdl.K_g: note.G,
	sdl.K_y: note.GSharp,
	sdl.K_h: note.A,
	sdl.K_u: note.ASharp,
	sdl.K_j: note.B,
	// high octave
	sdl.K_k: note.HC,
	sdl.K_o: note.HCSharp,
	sdl.K_l: note.HD,
}

var keyToCommand = map[sdl.Keycode]string{
	sdl.K_LEFTBRACKET:  "octave down",
	sdl.K_RIGHTBRACKET: "octave up",
}

func logKeyEvent(ev *sdl.KeyboardEvent) {
	fmt.Printf("[%d ms] Keyboard\ttype:%d\tsym:%c\tmodifiers:%d\tstate:%d\trepeat:%d\n",
		ev.Timestamp, ev.Type, ev.Keysym.Sym, ev.Keysym.Mod, ev.State, ev.Repeat)
}

func handleKeyEvent(ev *sdl.KeyboardEvent, wr *Writer) {
	kc := ev.Keysym.Sym

	noteModifier, notePressed := keyToNote[kc]
	command, commandPressed := keyToCommand[kc]

	if notePressed {
		absoluteNote := getAbsoluteNote(noteModifier)
		// first keydown = ev.State = 1, ev.Repeat = 0
		switch {
		case ev.State == 1 && ev.Repeat == 0:
			fmt.Println("pressed", absoluteNote)
			state.ActiveNotes[kc] = absoluteNote
			wr.NoteOn(uint8(absoluteNote), 90)
		case ev.State == 0:
			a, ok := state.ActiveNotes[kc]
			if ok {
				fmt.Println("released", a)
				delete(state.ActiveNotes, kc)
				wr.NoteOff(uint8(a))
			}
		}
	} else if commandPressed {
		if ev.State == 1 && ev.Repeat == 0 {
			switch command {
			case "octave down":
				if state.Octave > 2 {
					state.Octave -= 1
				}
			case "octave up":
				if state.Octave < 9 {
					state.Octave += 1
				}
			}
			fmt.Println(state)
		}
	} else {
		fmt.Printf("%d %d\n", sdl.GetScancodeFromKey(ev.Keysym.Sym), sdl.K_a)
	}
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

func draw(renderer *sdl.Renderer) {

	renderer.SetDrawColor(225, 225, 225, 255)
	renderer.Clear()

	var i, j int32
	var blackKeys []int32 = []int32{0, 1, 3, 4, 5}
	var rect sdl.Rect

	for i = 2; i < 10; i++ {
		r, g, b := getKeyboardColor(i)
		renderer.SetDrawColor(r, g, b, 255)
		kbOffset := 10 + 70*(i-2)
		rect = sdl.Rect{kbOffset, 10, 70, 40}

		renderer.FillRect(&rect)
		for j = 0; j < 8; j++ {
			renderer.SetDrawColor(50, 50, 50, 255)
			rect = sdl.Rect{kbOffset + j*10, 10, 10, 40}
			renderer.DrawRect(&rect)
		}
		// black keys
		for _, j := range blackKeys {
			renderer.SetDrawColor(50, 50, 50, 255)
			rect = sdl.Rect{kbOffset + 5 + j*10 + 2, 10, 6, 20}
			renderer.FillRect(&rect)
		}

		if i == int32(state.Octave) {
			renderer.SetDrawColor(255, 30, 30, 255)
			rect = sdl.Rect{kbOffset, 50, 90, 2}
			renderer.FillRect(&rect)
		}
	}

	renderer.SetDrawColor(255, 30, 30, 255)
	for _, abs := range state.ActiveNotes {
		oct := int32((abs - 24) / 12)
		var n note.NoteModifier = note.NoteModifier(abs % 12)

		kbOffset := (10 + oct*70)
		blackOffset, isBlack := blackOffsets[n]

		if isBlack {
			rect = sdl.Rect{kbOffset + blackOffset, 12, 4, 8}
		} else {
			rect = sdl.Rect{kbOffset + whiteOffsets[n], 40, 6, 8}
		}
		renderer.FillRect(&rect)
	}

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
				draw(renderer)
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
