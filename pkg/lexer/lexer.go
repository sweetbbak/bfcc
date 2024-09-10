package lexer

import (
	"strings"
)

// token types
const (
	EOF        = "EOF"
	DEC_PTR    = "<"
	INC_PTR    = ">"
	INC_CELL   = "+"
	DEC_CELL   = "-"
	OUTPUT     = "."
	INPUT      = ","
	LOOP_OPEN  = "["
	LOOP_CLOSE = "]"
)

type Token struct {
	// "<" "+" etc...
	Type string
	// number of consecutive tokens of this type
	Repeat int
}

type Lexer struct {
	// the BF program
	input string

	// the current position the lexer points to
	position int

	// map of characters to their token type
	known map[string]string

	// lets us determine if a token can have multiple occurences
	repeat map[string]bool
}

func New(input string) *Lexer {
	l := &Lexer{input: input}

	// clean up our input
	l.input = strings.ReplaceAll(l.input, "\n", "")
	l.input = strings.ReplaceAll(l.input, "\r", "")
	l.input = strings.ReplaceAll(l.input, " ", "")

	// register our known tokens
	l.known = make(map[string]string)
	l.known["+"] = INC_CELL
	l.known["-"] = DEC_CELL
	l.known[">"] = INC_PTR
	l.known["<"] = DEC_PTR
	l.known[","] = INPUT
	l.known["."] = OUTPUT
	l.known["["] = LOOP_OPEN
	l.known["]"] = LOOP_CLOSE

	// Some characters will have their input collapsed
	// when multiple consecutive occurrences are found.
	l.repeat = make(map[string]bool)
	l.repeat["+"] = true
	l.repeat["-"] = true
	l.repeat[">"] = true
	l.repeat["<"] = true

	return l
}

func Repl() *Lexer {
	l := &Lexer{}

	// register our known tokens
	l.known = make(map[string]string)
	l.known["+"] = INC_CELL
	l.known["-"] = DEC_CELL
	l.known[">"] = INC_PTR
	l.known["<"] = DEC_PTR
	l.known[","] = INPUT
	l.known["."] = OUTPUT
	l.known["["] = LOOP_OPEN
	l.known["]"] = LOOP_CLOSE

	// Some characters will have their input collapsed
	// when multiple consecutive occurrences are found.
	l.repeat = make(map[string]bool)
	l.repeat["+"] = true
	l.repeat["-"] = true
	l.repeat[">"] = true
	l.repeat["<"] = true

	return l
}

// takes an input string and returns its tokens
// overwrites lexers input, for use with the repl only.
func (l *Lexer) Read(inst string) []*Token {
	l.input = inst
	var res []*Token
	tok := l.Next()

	for tok.Type != EOF {
		res = append(res, tok)
		tok = l.Next()
	}

	return res
}

// reset the parsers position
func (l *Lexer) Zero() {
	l.position = 0
}

// returns all the tokens from the given input
func (l *Lexer) Tokens() []*Token {
	var res []*Token
	tok := l.Next()

	for tok.Type != EOF {
		res = append(res, tok)
		tok = l.Next()
	}

	return res
}

// advance the parser and get the next token in the input
// while counting repeated characters
func (l *Lexer) Next() *Token {
	// loop until the end
	for l.position < len(l.input) {
		char := string(l.input[l.position])

		// is this a valid token?
		_, ok := l.known[char]
		if ok {

			// can we repeat token
			repeatable := l.repeat[char]
			if !repeatable {
				l.position++
				return &Token{Type: char, Repeat: 1}
			}

			begin := l.position

			// if it is repeatable, we count how many times
			for l.position < len(l.input) {
				// if it isnt the same character, we are done
				if string(l.input[l.position]) != char {
					// chomp through repeated characters
					break
				}

				l.position++
			}

			// this gives us how many times this character was repeated
			count := l.position - begin
			return &Token{Type: char, Repeat: count}
		}
		// ignore unknown characters
		l.position++
	}

	// if we've made it here, we are done, send EOF
	return &Token{Type: EOF, Repeat: 1}
}
