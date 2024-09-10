package main

type Memfmt int

const (
	Hex Memfmt = iota
	Decimal
	Octal
)

type MemoryFormat struct {
	kind    Memfmt
	literal string
}

func (m *model) CycleMemFormat() {
	switch m.memfmt.kind {
	case Decimal:
		m.memfmt.kind = Hex
		m.memfmt.literal = "%d"
	case Hex:
		m.memfmt.kind = Octal
		m.memfmt.literal = "%x"
	case Octal:
		m.memfmt.kind = Decimal
		m.memfmt.literal = "%o"
	}
}
