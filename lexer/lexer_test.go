package lexer

import (
	"fmt"
	"testing"
)

func TestParse(t *testing.T) {
	input := "+++++[-]"
	l := New(input)

	tokens := l.Tokens()

	for _, t := range tokens {
		fmt.Printf("%s [%d]\n", t.Type, t.Repeat)
	}
}
