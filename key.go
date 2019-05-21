package main

import (
	"fmt"
	"github.com/minikomi/keyboye/internal/note"
	"github.com/veandco/go-sdl2/sdl"
)

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
	sdl.K_COMMA:  "octave down",
	sdl.K_PERIOD: "octave up",
}

func logKeyEvent(ev *sdl.KeyboardEvent) {
	fmt.Printf("[%d ms] Keyboard\ttype:%d\tsym:%c\tmodifiers:%d\tstate:%d\trepeat:%d\n",
		ev.Timestamp, ev.Type, ev.Keysym.Sym, ev.Keysym.Mod, ev.State, ev.Repeat)
}

func HandleKeyEvent(ev *sdl.KeyboardEvent, wr *Writer) {
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
