package main

import (
	"strconv"

	"github.com/minikomi/keyboye/internal/note"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

var srcRect, rect sdl.Rect
var i, j int32
var solid *sdl.Surface
var texture *sdl.Texture
var err error

var red sdl.Color = sdl.Color{225, 30, 30, 225}
var gray sdl.Color = sdl.Color{180, 180, 180, 225}

func Draw(renderer *sdl.Renderer, font *ttf.Font) {

	renderer.SetDrawColor(225, 225, 225, 255)
	renderer.Clear()

	// draw keyboard
	for i = 2; i < 10; i++ {
		kbOffsetLeft := 10 + 70*(i-2)

		// text
		if i == int32(state.Octave) {
			solid, err = font.RenderUTF8Solid(strconv.Itoa(int(i)), red)
		} else {
			solid, err = font.RenderUTF8Solid(strconv.Itoa(int(i)), gray)
		}
		if err != nil {
			panic(err.Error())
		}

		texture, err = renderer.CreateTextureFromSurface(solid)
		if err != nil {
			panic(err.Error())
		}

		srcRect := sdl.Rect{0, 0, 10, 12}
		rect = sdl.Rect{kbOffsetLeft, 0, 10, 12}
		renderer.Copy(texture, &srcRect, &rect)

		// bg
		r, g, b := getKeyboardColor(i)
		renderer.SetDrawColor(r, g, b, 255)
		rect = sdl.Rect{kbOffsetLeft, 12, 70, 40}
		renderer.FillRect(&rect)

		// keys
		for j = 0; j < 7; j++ {
			renderer.SetDrawColor(50, 50, 50, 255)
			rect = sdl.Rect{kbOffsetLeft + j*10, 12, 10, 40}
			renderer.DrawRect(&rect)
		}

		// black keys
		for _, j := range []int32{0, 1, 3, 4, 5} {
			renderer.SetDrawColor(50, 50, 50, 255)
			rect = sdl.Rect{kbOffsetLeft + 5 + j*10 + 2, 12, 6, 20}
			renderer.FillRect(&rect)
		}

		// active marker
		if i == int32(state.Octave) {
			renderer.SetDrawColor(255, 30, 30, 255)
			if i == 9 {
				rect = sdl.Rect{kbOffsetLeft, 52, 70, 2}
			} else {
				rect = sdl.Rect{kbOffsetLeft, 52, 90, 2}
			}
			renderer.FillRect(&rect)
		}
	}

	// draw pressed keys
	renderer.SetDrawColor(255, 30, 30, 255)
	for _, abs := range state.ActiveNotes {
		oct := int32((abs - 24) / 12)
		var n note.NoteModifier = note.NoteModifier(abs % 12)

		kbOffsetLeft := (10 + oct*70)
		blackOffset, isBlack := blackOffsets[n]

		if isBlack {
			rect = sdl.Rect{kbOffsetLeft + blackOffset, 12, 4, 8}
		} else {
			rect = sdl.Rect{kbOffsetLeft + whiteOffsets[n], 40, 6, 8}
		}
		renderer.FillRect(&rect)
	}

	renderer.Present()

}
