package main

import (
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
	TextColor2  lipgloss.Color
	TextField   lipgloss.Style
	TextField2  lipgloss.Style
	TextHelp    lipgloss.Style
	Border      lipgloss.Border
	Stdout      lipgloss.Style
}

func DefaultStyles() *Styles {
	s := new(Styles)
	s.BorderColor = lipgloss.Color("36")
	s.BorderBlur = lipgloss.Color("24")
	s.TextColor = lipgloss.Color("240")
	// s.TextColor = lipgloss.Color("24")
	s.TextColor2 = lipgloss.Color("33")
	s.TextField = lipgloss.NewStyle().BorderForeground(s.BorderColor).BorderStyle(lipgloss.DoubleBorder()).Foreground(s.TextColor)
	// s.TextField = lipgloss.NewStyle().BorderForeground(lipgloss.Color("69")).BorderStyle(lipgloss.DoubleBorder()).Foreground(s.TextColor2)
	s.Stdout = lipgloss.NewStyle().BorderForeground(s.BorderColor).BorderStyle(lipgloss.RoundedBorder())
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
	width        int
	height       int
	scroll       int
	input        textinput.Model
	fpOpen       bool
	styles       *Styles
	view         View // currently focused region
	vm           *debug.Debug
	memfmt       MemoryFormat // hex, octal, decimal memory layout
	content      string       // memory content
	step         *Stepper
	history      []string
	stdoutHeight int
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

	vm := debug.New(1200, true)

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
		return StdoutMsg(m.RenderStdout(m.width-3, m.stdoutHeight))
	})
}

type InstrMsg string

func (m model) UpdateBuffer() tea.Cmd {
	return tea.Tick(time.Microsecond, func(t time.Time) tea.Msg {
		return InstrMsg(m.RenderStdout(m.width-3, m.stdoutHeight))
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
			inputVal := m.input.Value()
			if strings.HasPrefix(inputVal, "open") {
				files := strings.Split(inputVal, " ")
				var file string
				if len(files) > 1 {
					file = files[1]
				} else {
					return m, nil
				}

				s, err := m.OpenFile(file)
				if err != nil {
					// render error out to tui?
					inputVal = err.Error()
				}

				if s != "" {
					inputVal = s
				}

				m.input.Reset()
				m.input.SetValue(inputVal)
				c := tea.Batch(m.UpdateEval(inputVal))
				return m, c
			}

			// only add valid brainfuck to the history
			m.input.Validate(inputVal)
			if m.input.Err == nil {
				m.history = append(m.history, inputVal)
				m.input.SetSuggestions(m.history)
			}

			m.input.Reset()
			c := tea.Batch(m.UpdateEval(inputVal))
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
func (m model) RenderStdout(maxwidth, height int) string {
	s := m.vm.SB.String()
	if s == "" {
		return ""
	}

	splits := strings.Split(s, "\n") // naive and slow probably

	for i, line := range splits {
		if len(line) > maxwidth {
			splits[i] = line[:maxwidth]
		}
	}

	if len(splits) < height {
		return s
	}

	lastN := splits[len(splits)-height:]
	return strings.Join(lastN, "\n")
}

// render the memory of the repl
func (m model) RenderState() string {
	s := m.vm.PrintState((m.width - 3) / 2)

	splits := strings.Split(s, "\n") // naive and slow probably
	if len(splits) < 5 {
		return s
	}

	lastFive := splits[len(splits)-5:]
	return strings.Join(lastFive, "\n")
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

	s += " ctrl+j speed++ | ctrl+k speed-- | reset | ctrl+a format | open <file>"

	return m.styles.TextHelp.Render(s)
}

func (m model) View() string {
	if m.width == 0 {
		return "loading..."
	}

	// input field
	answer := m.styles.TextField.
		Width(m.width - 2).
		Render(m.input.View())

	// help
	footer := m.RenderStatus()

	// instructions set
	outbuf := m.RenderState()
	outbuf = m.styles.Stdout.
		Width(m.width - 2).
		Height(2). // overflow
		Render(outbuf)

	// why -5 lmao is it the border again? Im guessing yes as its not necessarily accounted for ughh
	sheight := (m.height - lipgloss.Height(answer) - 5 - lipgloss.Height(footer) - lipgloss.Height(outbuf)) / 2
	m.stdoutHeight = sheight

	// emulated stdout
	stdout := m.RenderStdout(m.width-3, sheight)
	stdout = m.styles.Stdout.
		Width(m.width - 2).
		Height(sheight).
		Render(stdout)

	// mheight := m.height - lipgloss.Height(answer) - 2 - lipgloss.Height(footer) - lipgloss.Height(outbuf) - lipgloss.Height(stdout)

	// memory
	content := m.styles.TextField.
		Width(m.width - 2).
		Height(sheight - 1).
		Render(m.content)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		answer,
		content,
		stdout,
		outbuf,
		footer,
	)
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
