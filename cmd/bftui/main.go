package main

import (
	"fmt"
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
	width  int
	height int
	scroll int
	input  textinput.Model
	styles *Styles
	view   View // currently focused region
	vm     *debug.Debug
}

func initialModel() model {
	styles := DefaultStyles()
	input := textinput.New()
	input.Placeholder = "brainfuck instructions"
	input.Focus()
	inst := []string{"+", "-", "[", "]", ">", "<", ",", "."}
	input.SetSuggestions(inst)
	input.ShowSuggestions = true

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

	vm := debug.New(100, true)
	vm.SetStep(func() error {
		time.Sleep(time.Millisecond * 10)
		return nil
	})

	return model{
		styles: styles,
		input:  input,
		view:   0,
		vm:     vm,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
		case "enter":
			v := m.input.Value()
			m.input.Reset()
			// m.input.Placeholder = fmt.Sprintf("pointer %d", m.vm.Ptr())
			m.vm.Eval(v)
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
	// return m.vm.DumpMemory("%2d ", uint(m.width-3))
	return m.vm.DumpMemory(" %d ", uint(0))
}

func (m model) View() string {
	if m.width == 0 {
		return "loading..."
	}

	answer := m.styles.TextField.
		Width(m.width - 2).
		Render(m.input.View())

	mheight := m.height - lipgloss.Height(answer) - 2

	mem := m.RenderMemory()
	content := m.styles.TextField.
		Width(m.width - 2).
		Height(mheight).
		Render(mem)

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
