package note

type NoteModifier uint8

const (
	C = NoteModifier(iota)
	CSharp
	D
	DSharp
	E
	F
	FSharp
	G
	GSharp
	A
	ASharp
	B
	HC
	HCSharp
	HD
	HDSharp
	HE
)

type AbsoluteNote uint8
