// a special brainfuck stack that is concurrency safe
// and that allows you to step through instructions at any speed
// using a generated function. This is intended to operate like
// GDB or Blinkenlights but for brainfuck
package debug

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"bfcc/pkg/lexer"
	"github.com/muesli/reflow/wordwrap"
)

type Debug struct {
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
	// step function to regulate speed or instruction stepping
	step StepFn
	// current token
	curToken *lexer.Token
	// read write mutex
	rw sync.RWMutex
	// memory string colorer
	c Color
}

type StepFn func() error

// get a new interactive brainfuck repl
func New(stacksize int, hascolor bool) *Debug {
	l := lexer.Repl()

	vm := &Debug{
		Memory: make([]int, stacksize),
		ptr:    0,
		repl:   l,
		step:   func() error { return nil },
	}

	if hascolor {
		vm.c.Compute()
	}

	return vm
}

// return the current pointer value
func (v *Debug) SetStep(fn StepFn) {
	v.step = fn
}

// return the current pointer value
func (v *Debug) Ptr() int {
	return v.ptr
}

// print the current instruction set as a string
func (v *Debug) PrintState() string {
	var s string
	for i, t := range v.Tokens {
		if i == v.offset {
			s = fmt.Sprintf("\x1b[32m%s\x1b[0m", strings.Repeat(t.Type, t.Repeat))
		} else {
			s = fmt.Sprintf("%s", strings.Repeat(t.Type, t.Repeat))
		}
	}

	return s
}

// dump memory as a format "%d" or "%x"
// as well as a word wrap limit, use 0 to ignore
func (v *Debug) DumpMemory(format string, wrap int) string {
	var sb strings.Builder

	v.rw.RLock()
	defer v.rw.RUnlock()

	for i, n := range v.Memory {
		clr, nocolor := v.c.Colorize(byte(n))

		str := fmt.Sprintf(format, n)
		str2 := fmt.Sprintf("|%s%s%s", string(clr), str, string(nocolor))

		if v.ptr == i {
			str2 = fmt.Sprintf("\x1b[34m%s\x1b[0m", str2)
		}

		sb.WriteString(str2)
	}

	if wrap > 0 {
		return wordwrap.String(sb.String(), wrap)
	} else {
		return sb.String()
	}
}

// evaluate the given instruction. for use as a REPL
// we take instructions, tokenize them, and then modify the
// virtual machine structure accordingly.
func (v *Debug) Eval(instruction string) error {
	if v.repl == nil {
		return fmt.Errorf("repl has not been initialized")
	}

	tokens := v.repl.Read(instruction)
	v.Tokens = tokens
	v.offset = 0
	v.repl.Zero()

	for v.offset < len(v.Tokens) {
		// allow us to slow execution
		v.step()
		tok := v.Tokens[v.offset]
		v.curToken = tok

		// v.rw.Lock()
		// defer v.rw.Unlock()

		err := v.evaluate()
		if err != nil {
			return err
		}
	}

	return nil
}

// evaluate the current instruction
func (v *Debug) evaluate() error {
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
			v.step()
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
			v.step()
		}

		return nil
	}

	// next instruction
	v.offset++
	return nil
}
