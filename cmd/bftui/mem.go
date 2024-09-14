package main

import (
	"fmt"
	"strings"
)

type Memfmt int

const (
	Hex Memfmt = iota
	Decimal
	Char
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
		// m.memfmt.kind = Char
		m.memfmt.kind = Decimal
		m.memfmt.literal = "%x"
		// case Char:
		// 	m.memfmt.kind = Decimal
		// 	m.memfmt.literal = "%c"
	}
}

type Stepper struct {
	Speed   int       // speed of execution
	Step    chan bool // step channel user input -> step chan
	Running bool      // is stepping or is running
}

// pause execution
func (s *Stepper) Pause() {
	s.Running = false
	s.Speed = 0
}

// start execution with a default value of 10
func (s *Stepper) Run() {
	s.Running = true
	s.Speed = 10
}

// change the stepper speed at runtime
func (s *Stepper) ChangeSpeed(i int) {
	if s.Speed+i < 0 {
		s.Speed = 0
	} else {
		s.Speed += i
	}
}

func HighlightBF(s string) string {
	var sb strings.Builder
	for _, ch := range s {
		switch ch {
		case '>':
			sb.WriteString(fmt.Sprintf("\x1b[32m%s\x1b[0m", s))
		case '<':
			sb.WriteString(fmt.Sprintf("\x1b[33m%s\x1b[0m", s))
		case '+':
			sb.WriteString(fmt.Sprintf("\x1b[34m%s\x1b[0m", s))
		case '-':
			sb.WriteString(fmt.Sprintf("\x1b[35m%s\x1b[0m", s))
		case '[':
			sb.WriteString(fmt.Sprintf("\x1b[36m%s\x1b[0m", s))
		case ']':
			sb.WriteString(fmt.Sprintf("\x1b[36m%s\x1b[0m", s))
		case ',':
			sb.WriteString(fmt.Sprintf("\x1b[37m%s\x1b[0m", s))
		case '.':
			sb.WriteString(fmt.Sprintf("\x1b[38m%s\x1b[0m", s))
		default:
			sb.WriteRune(ch)
		}
	}
	return sb.String()
}
