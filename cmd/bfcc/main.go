package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"bfcc/pkg/gen/c"
	"bfcc/pkg/gen/interp"
	"bfcc/pkg/repl"
	"github.com/jessevdk/go-flags"
)

type Options struct {
	Output    string `short:"o" long:"output" description:"binary executable to output to"`
	Run       bool   `short:"r" long:"run" description:"run executable after compiling"`
	Repl      bool   `short:"R" long:"repl" description:"run the interactive brainfuck interpreter"`
	Backend   string `short:"b" long:"backend" description:"what backend to use [C, ASM, Native, VM]"`
	StackSize uint   `short:"s" long:"stack-size" description:"how much 'memory' to use"`
	Input     string `short:"i" long:"input" description:"input brainfuck file"`
}

var opts Options

func init() {
	opts.StackSize = 30_000
	opts.Backend = "c"
	opts.Output = "a.out"
}

func Interp(input string) error {
	vm := interp.New(int(opts.StackSize))
	vm.Input = os.Stdin
	vm.Output = os.Stdout

	return vm.Generate(input, opts.Output)
}

func Run(args []string) error {
	var input string

	if opts.Input != "" {
		input = opts.Input
	} else if len(args) > 0 {
		input = args[0]
	} else {
		return fmt.Errorf("no input source file provided")
	}

	b, err := os.ReadFile(input)
	if err != nil {
		return err
	}

	// run interpreter
	if opts.Backend[0] == 'i' || opts.Backend == "vm" {
		return Interp(string(b))
	}

	cgen := cgen.New(opts.StackSize)

	err = cgen.Generate(string(b), opts.Output)
	if err != nil {
		return err
	}

	if opts.Run {
		cmdname, err := filepath.Abs(opts.Output)
		if err != nil {
			return err
		}

		exe := exec.Command(cmdname)
		exe.Stdin = os.Stdin
		exe.Stdout = os.Stdout
		exe.Stderr = os.Stderr

		err = exe.Run()
		if err != nil {
			return fmt.Errorf("Error launching %s: %s\n", opts.Output, err)
		}
	}

	return nil
}

func RunRepl() error {
	// return repl.Start()
	return repl.Readline()
}

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		if flags.WroteHelp(err) {
			os.Exit(0)
		} else {
			log.Fatal(err)
		}
	}

	if opts.Repl {
		if err := RunRepl(); err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}

	if err := Run(args); err != nil {
		log.Fatal(err)
	}
}
