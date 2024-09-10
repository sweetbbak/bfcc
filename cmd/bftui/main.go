package main

import (
	"log"
	"time"

	debug "bfcc/dbg"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Styles struct {
	BorderColor lipgloss.Color
	BorderBlur  lipgloss.Color
	TextColor   lipgloss.Color
	TextField   lipgloss.Style
	Border      lipgloss.Border
}

func DefaultStyles() *Styles {
	s := new(Styles)
	s.BorderColor = lipgloss.Color("36")
	s.BorderBlur = lipgloss.Color("240")
	s.TextColor = lipgloss.Color("240")
	s.TextField = lipgloss.NewStyle().BorderForeground(s.BorderColor).BorderStyle(lipgloss.RoundedBorder()).Foreground(s.TextColor)
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
}

func initialModel() model {
	styles := DefaultStyles()
	input := textinput.New()
	input.Placeholder = "brainfuck instructions"
	input.Focus()
	inst := []string{"+", "-", "[", "]", ">", "<", ",", "."}
	input.SetSuggestions(inst)
	input.ShowSuggestions = true
	input.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	input.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	m := MemoryFormat{}
	m.kind = Decimal
	m.literal = "%d"

	// input.Validate = func(s string) error {
	// 	for _, c := range s {
	// 		switch string(c) {
	// 		case "+", "-", "[", "]", ">", "<", ",", ".":
	// 		default:
	// 			return fmt.Errorf("invalid instruction")
	// 		}
	// 	}
	// 	return nil
	// }

	vm := debug.New(150, true)

	vm.SetStep(func() error {
		time.Sleep(time.Millisecond * 10)
		return nil
	})

	return model{
		styles: styles,
		input:  input,
		view:   0,
		vm:     vm,
		memfmt: m,
	}
}

func (m model) Init() tea.Cmd {
	return m.UpdateEval2()
}

type EvalMsg error

func (m model) UpdateEval(input string) tea.Cmd {
	return func() tea.Msg {
		err := m.vm.Eval(input)
		return EvalMsg(err)
	}
}

type EvalMsgx struct {
	content string
	t       time.Time
}

func (m model) UpdateEval2() tea.Cmd {
	return tea.Tick(time.Microsecond, func(t time.Time) tea.Msg {
		obj := EvalMsgx{
			content: m.RenderMemory(),
			t:       t,
		}
		return EvalMsgx(obj)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case EvalMsgx:
		m.content = msg.content
		return m, m.UpdateEval2()
	case EvalMsg:
		if msg != nil {
			m.input.Placeholder = msg.Error()
		}
		return m, m.UpdateEval2()
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "ctrl+a":
			m.CycleMemFormat()
		case "tab":
		case "enter":
			v := m.input.Value()
			m.input.Reset()
			c := tea.Batch(m.UpdateEval(v))
			return m, c
		case "esc", "escape":
		}
	}

	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// change the models focus
func (m model) CycleZone() tea.Cmd {
	m.view = (m.view + 1) % m.view
	return nil
}

// render the memory of the repl
func (m model) RenderMemory() string {
	return m.vm.DumpMemory(m.memfmt.literal, m.width/2)
}

func (m model) View() string {
	if m.width == 0 {
		return "loading..."
	}

	answer := m.styles.TextField.
		Width(m.width - 2).
		Render(m.input.View())

	mheight := m.height - lipgloss.Height(answer) - 2

	content := m.styles.TextField.
		Width(m.width - 2).
		Height(mheight).
		Render(m.content)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		answer,
		content,
	)
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
