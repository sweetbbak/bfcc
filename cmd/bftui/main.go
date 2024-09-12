package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	debug "bfcc/pkg/dbg"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Styles struct {
	BorderColor lipgloss.Color
	BorderBlur  lipgloss.Color
	TextColor   lipgloss.Color
	TextField   lipgloss.Style
	TextHelp    lipgloss.Style
	Border      lipgloss.Border
}

func DefaultStyles() *Styles {
	s := new(Styles)
	s.BorderColor = lipgloss.Color("36")
	s.BorderBlur = lipgloss.Color("240")
	s.TextColor = lipgloss.Color("240")
	s.TextField = lipgloss.NewStyle().BorderForeground(s.BorderColor).BorderStyle(lipgloss.RoundedBorder()).Foreground(s.TextColor)
	s.TextHelp = lipgloss.NewStyle().Foreground(s.TextColor)
	s.Border = lipgloss.RoundedBorder()
	return s
}

type View int

const (
	input View = iota
	memory
	instructions
)

type model struct {
	width   int
	height  int
	scroll  int
	input   textinput.Model
	styles  *Styles
	view    View // currently focused region
	vm      *debug.Debug
	memfmt  MemoryFormat // hex, octal, decimal memory layout
	content string       // memory content
	step    *Stepper
	history []string
	xoutput *bytes.Buffer
	output  *bufio.Writer
}

func initialModel() model {
	styles := DefaultStyles()

	input := textinput.New()
	input.Focus()
	input.Placeholder = "brainfuck"

	inst := []string{"+", "-", "[", "]", ">", "<", ",", ".", "pause", "help"}
	input.SetSuggestions(inst)
	input.ShowSuggestions = true

	input.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	input.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	m := MemoryFormat{}
	m.kind = Decimal
	m.literal = "%d"

	input.Validate = func(s string) error {
		for _, c := range s {
			switch string(c) {
			case "+", "-", "[", "]", ">", "<", ",", ".":
			default:
				return fmt.Errorf("invalid instruction")
			}
		}
		return nil
	}

	vm := debug.New(200, true)

	// emulate stdout
	// var outbuf *bytes.Buffer
	// outbuf := new(bytes.Buffer)
	var outbuff strings.Builder
	// w := bufio.NewWriter(outbuf)
	// r := bufio.NewReader(os.Stdin)
	// rw := bufio.NewReadWriter(r, w)

	// bytes.Buffer not showing up?
	// vm.Output = outbuf
	vm.Output = &outbuff
	// vm.Output = w
	// vm.Output = os.Stdout
	vm.Input = os.Stdin

	vm.SetStep(func() error {
		time.Sleep(time.Millisecond * 10)
		return nil
	})

	s := &Stepper{
		Speed:   10,
		Step:    make(chan bool, 1),
		Running: true,
	}

	return model{
		styles: styles,
		input:  input,
		view:   0,
		vm:     vm,
		memfmt: m,
		step:   s,
	}
}

func (m model) Init() tea.Cmd {
	cmd := tea.Batch(m.UpdateStdout(), m.UpdateMemory())
	return cmd
	// return m.UpdateMemory()
}

type EvalMsg error

func (m model) UpdateEval(input string) tea.Cmd {
	return func() tea.Msg {
		err := m.vm.Eval(input)
		return EvalMsg(err)
	}
}

type MemoryMsg struct {
	content string
	t       time.Time
}

func (m model) UpdateMemory() tea.Cmd {
	return tea.Tick(time.Microsecond, func(t time.Time) tea.Msg {
		obj := MemoryMsg{
			content: m.RenderMemory(),
			t:       t,
		}
		return MemoryMsg(obj)
	})
}

type StdoutMsg string

func (m model) UpdateStdout() tea.Cmd {
	return tea.Tick(time.Microsecond, func(t time.Time) tea.Msg {
		// return StdoutMsg(m.RenderStdout())
		// m.output.Flush()
		// log.Println(m.xoutput.String())
		return StdoutMsg(m.vm.SB.String())
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case MemoryMsg:
		m.content = msg.content
		return m, m.UpdateMemory()
	case StdoutMsg:
		log.Println(msg)
		return m, m.UpdateStdout()
	case EvalMsg:
		if msg != nil {
			m.input.Placeholder = msg.Error()
		}
		return m, m.UpdateMemory()
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "ctrl+a":
			m.CycleMemFormat()
		case "tab":
		case "ctrl+j":
			m.step.ChangeSpeed(-2) // ironically this speeds things up lol
			m.vm.SetStep(func() error {
				time.Sleep(time.Millisecond * time.Duration(m.step.Speed))
				return nil
			})
		case "ctrl+k":
			m.step.ChangeSpeed(2)
			m.vm.SetStep(func() error {
				time.Sleep(time.Millisecond * time.Duration(m.step.Speed))
				return nil
			})
		case "ctrl+p":
			if m.step.Running {
				// not working lol
				m.step.Pause()
				m.vm.SetStep(func() error {
					<-m.step.Step
					return nil
				})
			} else {
				m.step.Run()
				m.vm.SetStep(func() error {
					time.Sleep(time.Millisecond * time.Duration(m.step.Speed))
					return nil
				})
			}
		case "ctrl+s":
			if !m.step.Running {
				m.step.Step <- true
			}
		case "enter":
			v := m.input.Value()
			if len(v) > len("open") && strings.HasPrefix(v, "open") {
				files := strings.Split(v, " ")
				var file string
				if len(files) > 1 {
					file = files[1]
				} else {
					return m, nil
				}

				s, err := m.OpenFile(file)
				if err != nil {
					// render error out to tui?
					v = err.Error()
				}

				if s != "" {
					v = s
				} else {
					v = "file is empty"
				}

				m.input.Reset()
				m.input.SetValue(v)
				c := tea.Batch(m.UpdateEval(v))
				return m, c
			}

			// only add valid brainfuck to the history
			m.input.Validate(v)
			if m.input.Err == nil {
				m.history = append(m.history, v)
				m.input.SetSuggestions(m.history)
			}

			m.input.Reset()
			c := tea.Batch(m.UpdateEval(v))
			return m, c
		case "esc", "escape":
		}
	}

	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) OpenFile(file string) (string, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// change the models focus
func (m model) CycleZone() tea.Cmd {
	m.view = (m.view + 1) % m.view
	return nil
}

// render the memory of the repl
func (m model) RenderStdout() string {
	x := m.vm.PrintState()
	y := m.vm.SB.String()
	return fmt.Sprintf("%s\n%s", x, y)
}

// render the memory of the repl
func (m model) RenderMemory() string {
	return m.vm.DumpMemory(m.memfmt.literal, m.width/2)
}

func (m model) RenderStatus() string {
	var s string
	if m.step.Running {
		s = fmt.Sprintf("running: speed %d |", m.step.Speed)
	} else {
		s += "paused"
	}

	s += " ctrl+j speed++ | ctrl+k speed--"

	return m.styles.TextHelp.Render(s)
}

func (m model) View() string {
	if m.width == 0 {
		return "loading..."
	}

	answer := m.styles.TextField.
		Width(m.width - 2).
		Render(m.input.View())

	// outbuf := m.styles.TextField.Render(m.RenderStdout())
	outbuf := m.RenderStdout()
	footer := m.RenderStatus()
	mheight := m.height - lipgloss.Height(answer) - 2 - lipgloss.Height(footer) - lipgloss.Height(outbuf)

	content := m.styles.TextField.
		Width(m.width - 2).
		Height(mheight).
		Render(m.content)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		answer,
		content,
		outbuf,
		footer,
	)
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())

	f, _ := os.Create("tmp.log")
	log.SetOutput(f)

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
