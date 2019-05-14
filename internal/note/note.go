package note

type NoteModifier uint8

const (
	C = NoteModifier(iota)
	CSharp
	D
	DSharp
	E
	F
	G
	GSharp
	A
	ASharp
	B
	HC
	HCSharp
	HD
	HDSharp
)

type AbsoluteNote uint8
