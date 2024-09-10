package repl

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"

	"bfcc/pkg/gen/interp"
)

func Start() error {
	repl := interp.NewRepl(100)
	const prompt = "\x1b[32m[bf]\x1b[0m \x1b[34m~ $\x1b[0m "

	var outbuf bytes.Buffer
	// w := bufio.NewWriter(&outbuf)

	repl.Output = os.Stdout
	// cant do this find fix
	repl.Input = os.Stdin
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println(repl.Memory)
		fmt.Print(prompt)

		scan := scanner.Scan()
		if !scan {
			return fmt.Errorf("Error reading input")
		}

		line := scanner.Text()
		cmd := strings.Split(line, " ")

		switch cmd[0] {
		case "exit":
			return nil
		case "help":
			fmt.Println("help, exit, ptr, buf, clear, open <file> and brainfuck operators '+-<>[].,'")
			continue
		case "pointer", "ptr":
			fmt.Printf("ptr value: %d\n", repl.Ptr())
		case "buf":
			fmt.Printf("%s\n", outbuf.String())
		case "clear":
			fmt.Print("\x1b[2J\x1b[H")
		case "open":
			if len(cmd) < 2 {
				fmt.Println("no file provided")
				continue
			}

			b, err := os.ReadFile(cmd[1])
			if err != nil {
				log.Println(err)
			}

			repl.Eval(string(b))
			continue

		default:
			repl.Eval(line)
		}
	}
}
