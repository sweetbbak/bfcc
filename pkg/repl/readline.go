package repl

import (
	"fmt"
	"log"
	"os"
	"strings"

	"bfcc/pkg/gen/interp"
	repl "github.com/openengineer/go-repl"
)

type Bhandler struct {
	prompt string
	r      *repl.Repl
	rpl    *interp.Interpreter
}

func (h *Bhandler) Prompt() string {
	return h.prompt
}

func (h *Bhandler) Tab(buffer string) string {
	// a tab is simply 2 spaces here
	if len(buffer) < 1 {
		return "  "
	}

	if buffer[0] == 'h' {
		return "elp"
	}

	return "  "
}

// return is for shell history
func (h *Bhandler) Eval(buffer string) string {
	fmt.Println(h.rpl.Memory)

	// upon eval the Stdin should be unblocked
	if strings.TrimSpace(buffer) != "" {
		if buffer == "quit" || buffer == "exit" {
			h.r.Quit()
			return ""
		}

		fields := strings.Fields(buffer)
		cmd, args := fields[0], fields[1:]

		switch cmd {
		case "help":
			// fmt.Println("help, exit, ptr, buf, clear, open <file> and brainfuck operators '+-<>[].,'")
			return "\rhelp, exit, ptr, buf, clear, open <file> and brainfuck operators '+-<>[].,'"
		case "pointer", "ptr":
			return fmt.Sprintf("ptr value: %d\n", h.rpl.Ptr())
		case "buf":
		case "clear":
			fmt.Print("\x1b[2J\x1b[H")
		case "open":
			if len(cmd) < 2 {
				return "no file provided"
			}

			b, err := os.ReadFile(args[0])
			if err != nil {
				log.Print(err)
			}

			h.rpl.Eval(string(b))
		default:
			h.rpl.Eval(buffer)
		}

	} else {
		return ""
	}

	return ""
}

func Readline() error {
	h := &Bhandler{}
	h.r = repl.NewRepl(h)
	const prompt = "\x1b[32m[bf]\x1b[0m \x1b[34m~ $\x1b[0m "
	h.prompt = prompt
	h.prompt = "~$ "

	rpl := interp.NewRepl(100)
	rpl.Output = os.Stdout
	rpl.Input = os.Stdin

	h.rpl = rpl

	if err := h.r.Loop(); err != nil {
		return fmt.Errorf("%s\n", err.Error())
	}

	return nil
}
