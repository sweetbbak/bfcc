# bfcc

An over egineered `brainf*ck` compiler and interactive debugger written in Golang and Bubbletea.

![example of bftui running](assets/bftui.gif)

# Installation

you can use `just` to build:

```sh
# build the compiler
just
# build the TUI
just bftui
```

or

```sh
go build -o bfcc ./cmd/bfcc
go build -o bftui ./cmd/bftui

# or install with Go
go install github.com/sweetbbak/bfcc/cmd/bfcc@latest
go install github.com/sweetbbak/bfcc/cmd/bftui@latest
```

# Usage

```sh
# building a simple program (uses C backend by default)
./bfcc ./examples/helloworld.bf -o hello
# run the live interpreter (you can shorthand interpreter as interp or w/e as long as the first char is 'i')
./bfcc --backend=interpreter ./examples/helloworld.bf
# optionally execute the compiled program
./bfcc --backend=go ./examples/helloworld.bf -o hello --run
```

running the debugger UI:

```sh
./bftui
```

## backends

- Go
- C
- Asm
- Interpreted

### Benchmarks

| Command                                            |       Mean [ms] | Min [ms] | Max [ms] |     Relative |
| :------------------------------------------------- | --------------: | -------: | -------: | -----------: |
| `./mandelbrot-c`                                   |     512.0 ± 1.7 |    510.3 |    514.6 |         1.00 |
| `./mandelbrot-go`                                  |    2217.8 ± 1.2 |   2216.1 |   2219.6 |  4.33 ± 0.01 |
| `./bfcc --backend=interp ./examples/mandelbrot.bf` | 13572.8 ± 154.1 |  13347.0 |  13808.8 | 26.51 ± 0.31 |

# Huge thanks to

![Katie Ball](https://gist.github.com/roachhd/dce54bec8ba55fb17d3a)
![skx/bfcc](https://github.com/skx/bfcc)
![brainfuck.org](http://brainfuck.org)
