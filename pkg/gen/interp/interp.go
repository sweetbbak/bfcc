package interp

import (
	"fmt"
	"io"

	"bfcc/pkg/lexer"
)

type Interpreter struct {
	// the programs tokens
	Tokens []*lexer.Token
	// out programs memory / tape
	Memory []int
	// usually stdin, for ',' read instruction
	Input io.Reader
	// usually stdout, for writing to
	Output io.Writer
	// our position in the tokens
	offset int
	// brainfuck pointer
	ptr int
	// repl
	repl *lexer.Lexer
}

// get a new interactive brainfuck Virtual Machine
func New(stacksize int) *Interpreter {
	vm := &Interpreter{
		Memory: make([]int, stacksize),
		ptr:    0,
	}

	return vm
}

// interpret an entire brainfuck program
func (v *Interpreter) Generate(input string, output string) error {
	l := lexer.New(input)

	tok := l.Next()
	for tok.Type != lexer.EOF {
		v.Tokens = append(v.Tokens, tok)
		tok = l.Next()
	}

	v.ptr = 0
	v.offset = 0

	for v.offset < len(v.Tokens) {
		err := v.evaluate()
		if err != nil {
			return err
		}
	}

	return nil
}

// get a new interactive brainfuck repl
func NewRepl(stacksize int) *Interpreter {
	l := lexer.Repl()

	vm := &Interpreter{
		Memory: make([]int, stacksize),
		ptr:    0,
		repl:   l,
	}

	return vm
}

// return the current pointer value
func (v *Interpreter) Ptr() int {
	return v.ptr
}

// evaluate the given instruction. for use as a REPL
// we take instructions, tokenize them, and then modify the
// virtual machine structure accordingly.
func (v *Interpreter) Eval(instruction string) error {
	if v.repl == nil {
		return fmt.Errorf("repl has not been initialized")
	}

	tokens := v.repl.Read(instruction)
	v.Tokens = tokens
	v.offset = 0
	v.repl.Zero()

	for v.offset < len(v.Tokens) {
		err := v.evaluate()
		if err != nil {
			return err
		}
	}

	return nil
}

// evaluate the current instruction
func (v *Interpreter) evaluate() error {
	tok := v.Tokens[v.offset]
	switch tok.Type {

	case lexer.INC_PTR:
		v.ptr += tok.Repeat

	case lexer.DEC_PTR:
		v.ptr -= tok.Repeat

	case lexer.INC_CELL:
		v.Memory[v.ptr] += tok.Repeat

	case lexer.DEC_CELL:
		v.Memory[v.ptr] -= tok.Repeat

	case lexer.OUTPUT:
		fmt.Fprintf(v.Output, "%c", rune(v.Memory[v.ptr]))

	case lexer.INPUT:
		buf := make([]byte, 1)
		b, err := v.Input.Read(buf)
		if err != nil {
			return err
		}

		if b != 1 {
			return fmt.Errorf("read %d bytes of input, not 1", b)
		}

		v.Memory[v.ptr] = int(buf[0])

	case lexer.LOOP_OPEN:
		// advance if our loop counter is not 0 (which is when we stop looping)
		if v.Memory[v.ptr] != 0 {
			v.offset++
			return nil
		}

		// counting the depth of our loops (they are often deeply nested)
		depth := 1
		for depth != 0 {
			v.offset++
			switch v.Tokens[v.offset].Type {
			case lexer.LOOP_OPEN:
				depth++
			case lexer.LOOP_CLOSE:
				depth--
			}
		}

		return nil

	case lexer.LOOP_CLOSE:
		// if our loop is over, move on
		if v.Memory[v.ptr] == 0 {
			v.offset++
			return nil
		}

		// counting the depth of our loops (they are often deeply nested)
		// same as LOOP_OPEN but in reverse
		depth := 1
		for depth != 0 {
			v.offset--
			switch v.Tokens[v.offset].Type {
			case lexer.LOOP_OPEN:
				depth--
			case lexer.LOOP_CLOSE:
				depth++
			}
		}

		return nil
	}

	// next instruction
	v.offset++
	return nil
}
