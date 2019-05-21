package main

import (
	"fmt"
	"github.com/gomidi/connect"
	"github.com/gomidi/midi"
	"github.com/gomidi/midi/midimessage/channel"
	"github.com/gomidi/midi/midiwriter"
	"io"
)

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

func CreateMidiWriterTo(out connect.Out) *Writer {
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

func PrintPort(port connect.Port) {
	fmt.Printf("[%v] %s\n", port.Number(), port.String())
}

func PrintInPorts(ports []connect.In) {
	fmt.Printf("MIDI IN Ports\n")
	for _, port := range ports {
		PrintPort(port)
	}
	fmt.Printf("\n\n")
}

func PrintOutPorts(ports []connect.Out) {
	fmt.Printf("MIDI OUT Ports\n")
	for _, port := range ports {
		PrintPort(port)
	}
	fmt.Printf("\n\n")
}
