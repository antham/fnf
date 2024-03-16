package component

import (
	"errors"
	"log"
	"slices"
	"strings"

	"github.com/antham/fnf/forward"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.design/x/clipboard"
	"golang.org/x/term"
)

type state int

const (
	forwardListState state = iota
	inputState
	errorState
)

var (
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle         = focusedStyle.Copy()
	noStyle             = lipgloss.NewStyle()
	helpStyle           = blurredStyle.Copy()
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
)

type ForwardList struct {
	forward      *forward.OVHProvider
	table        table.Model
	defaultEmail string
	state        state
	inputs       []textinput.Model
	inputSwitch  int
}

func NewForwardList(forward *forward.OVHProvider) (tea.Model, error) {
	width, _, err := term.GetSize(0)
	if err != nil {
		return nil, err
	}
	columns := []table.Column{
		{Title: "Forward email", Width: width / 2},
		{Title: "Destination", Width: width / 2},
	}
	list, err := forward.List()
	if err != nil {
		return nil, err
	}
	rows := createRows(list)
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)
	t.SetStyles(s)
	prefixField := textinput.New()
	prefixField.Placeholder = "Email prefix"
	prefixField.CharLimit = 30
	forwardEmailField := textinput.New()
	forwardEmailField.Placeholder = "Forward email"
	forwardEmailField.CharLimit = 30
	inputs := []textinput.Model{prefixField, forwardEmailField}
	return ForwardList{forward: forward, table: t, state: forwardListState, inputs: inputs}, nil
}

func createRows(list []forward.ForwardInfo) []table.Row {
	rows := []table.Row{}
	for _, l := range list {
		rows = append(rows, table.Row{l.From, l.To})
	}
	return rows
}

func (m ForwardList) Init() tea.Cmd { return nil }

func (m ForwardList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		k := msg.String()

		if k == "q" || k == "ctrl+c" {
			return m, tea.Quit
		}
	}
	switch m.state {
	case forwardListState:
		var cmd tea.Cmd
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				if m.table.Focused() {
					m.table.Blur()
				} else {
					m.table.Focus()
				}
			case "q", "ctrl+c":
				return m, tea.Quit
			case "enter":
			case "n":
				err := m.forward.CreateOnDefaultEmail()
				if err != nil {
					log.Fatal(err)
				}
				list, err := m.forward.List()
				if err != nil {
					log.Fatal(err)
				}
				m.table.SetRows(createRows(list))
			case "c":
				clipboard.Write(clipboard.FmtText, []byte(m.table.SelectedRow()[1]))
			case "a":
				m.state = inputState
			case "d":
				list, err := m.forward.List()
				if err != nil {
					log.Fatal(err)
				}
				i := slices.IndexFunc(list, func(e forward.ForwardInfo) bool {
					return m.table.SelectedRow()[0] == e.From
				})
				if i == -1 {
					log.Fatal(errors.New("entry cannot be deleted"))
				}
				if err := m.forward.Delete(list[i].ID); err != nil {
					log.Fatal(err)
				}
				list = slices.DeleteFunc(list, func(e forward.ForwardInfo) bool {
					return m.table.SelectedRow()[0] == e.From
				})
				m.table.SetRows(createRows(list))
			}
		}
		m.table, cmd = m.table.Update(msg)
		return m, cmd
	case inputState:
		var cmd tea.Cmd
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				err := m.forward.Create(m.inputs[0].Value(), m.inputs[1].Value())
				if err != nil {
					log.Fatal(err)
				}
				list, err := m.forward.List()
				if err != nil {
					log.Fatal(err)
				}
				m.table.SetRows(createRows(list))
				m.inputs[0].Reset()
				m.inputs[1].Reset()
				m.state = forwardListState
			case "tab", "up", "down":
				if m.inputSwitch == 1 {
					m.inputSwitch = 0
					break
				}
				m.inputSwitch += 1
			case "esc":
				m.state = forwardListState
			}
		}
		cmds := make([]tea.Cmd, len(m.inputs))
		for i := 0; i <= len(m.inputs)-1; i++ {
			if i == m.inputSwitch {
				cmds[i] = m.inputs[i].Focus()
				m.inputs[i].PromptStyle = focusedStyle
				m.inputs[i].TextStyle = focusedStyle
				continue
			}
			m.inputs[i].Blur()
			m.inputs[i].PromptStyle = noStyle
			m.inputs[i].TextStyle = noStyle
			m.inputs[m.inputSwitch].Cursor.SetMode(cursor.CursorBlink)
			m.inputs[m.inputSwitch].Cursor.Style = cursorStyle
		}
		m.inputs[m.inputSwitch], cmd = m.inputs[m.inputSwitch].Update(msg)
		return m, cmd
	}
	return m, tea.Quit
}

func (m ForwardList) View() string {
	v := lipgloss.NewStyle().MarginTop(1)
	var view string
	switch m.state {
	case errorState:

	case inputState:
		var b strings.Builder
		for i := range m.inputs {
			b.WriteString(m.inputs[i].View())
			if i < len(m.inputs)-1 {
				b.WriteRune('\n')
			}
		}
		view = b.String()
	case forwardListState:
		view = m.table.View()
	}
	return v.Render(view)
}
