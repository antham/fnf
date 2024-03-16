package component

import (
	"errors"
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
	cursorStyle         = focusedStyle
	noStyle             = lipgloss.NewStyle()
	helpStyle           = blurredStyle
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
)

type ForwardList struct {
	forward       *forward.OVHProvider
	tableModel    table.Model
	defaultEmail  string
	state         state
	editMailModel []textinput.Model
	inputSwitch   int
	err           error
	windowWidth   int
}

func NewForwardList(forward *forward.OVHProvider, defaultEmail string) (tea.Model, error) {
	width, height, err := term.GetSize(0)
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
		table.WithHeight(height-1),
	)
	t.SetStyles(s)
	prefixField := textinput.New()
	prefixField.Placeholder = "Email prefix"
	prefixField.CharLimit = 30
	forwardEmailField := textinput.New()
	forwardEmailField.Placeholder = "Forward email"
	forwardEmailField.CharLimit = 30
	forwardEmailField.SetValue(defaultEmail)
	inputs := []textinput.Model{prefixField, forwardEmailField}
	return ForwardList{forward: forward, tableModel: t, state: forwardListState, defaultEmail: defaultEmail, editMailModel: inputs, windowWidth: width}, nil
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
	case errorState:
		var cmd tea.Cmd
		switch msg.(type) {
		case tea.WindowSizeMsg:
		case tea.KeyMsg:
			m.state = forwardListState
			return m, cmd
		}
	case forwardListState:
		var cmd tea.Cmd
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				if m.tableModel.Focused() {
					m.tableModel.Blur()
				} else {
					m.tableModel.Focus()
				}
			case "q", "ctrl+c":
				return m, tea.Quit
			case "enter":
			case "n":
				err := m.forward.CreateOnDefaultEmail()
				if err != nil {
					m.err = err
					m.state = errorState
					return m, cmd
				}
				list, err := m.forward.List()
				if err != nil {
					m.err = err
					m.state = errorState
					return m, cmd
				}
				m.tableModel.SetRows(createRows(list))
				m.tableModel.GotoTop()
			case "c":
				clipboard.Write(clipboard.FmtText, []byte(m.tableModel.SelectedRow()[1]))
			case "a":
				m.state = inputState
			case "d":
				list, err := m.forward.List()
				if err != nil {
					m.err = err
					m.state = errorState
					return m, cmd
				}
				i := slices.IndexFunc(list, func(e forward.ForwardInfo) bool {
					return m.tableModel.SelectedRow()[0] == e.From
				})
				if i == -1 {
					m.err = errors.New("entry cannot be deleted")
					m.state = errorState
					return m, cmd
				}
				err = m.forward.Delete(list[i].ID)
				if err != nil {
					m.err = err
					m.state = errorState
					return m, cmd
				}
				list = slices.DeleteFunc(list, func(e forward.ForwardInfo) bool {
					return m.tableModel.SelectedRow()[0] == e.From
				})
				m.tableModel.SetRows(createRows(list))
				m.tableModel.GotoTop()
			}
		}
		m.tableModel, cmd = m.tableModel.Update(msg)
		return m, cmd
	case inputState:
		var cmd tea.Cmd
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				err := m.forward.Create(strings.TrimSpace(m.editMailModel[0].Value()), strings.TrimSpace(m.editMailModel[1].Value()))
				if err != nil {
					m.err = err
					m.state = errorState
					return m, cmd
				}
				list, err := m.forward.List()
				if err != nil {
					m.err = err
					m.state = errorState
					return m, cmd
				}
				m.tableModel.SetRows(createRows(list))
				m.editMailModel[0].Reset()
				m.editMailModel[1].SetValue(m.defaultEmail)
				m.tableModel.GotoTop()
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
		cmds := make([]tea.Cmd, len(m.editMailModel))
		for i := 0; i < len(m.editMailModel); i++ {
			if i == m.inputSwitch {
				cmds[i] = m.editMailModel[i].Focus()
				m.editMailModel[i].PromptStyle = focusedStyle
				m.editMailModel[i].TextStyle = focusedStyle
				continue
			}
			m.editMailModel[i].Blur()
			m.editMailModel[i].PromptStyle = noStyle
			m.editMailModel[i].TextStyle = noStyle
			m.editMailModel[m.inputSwitch].Cursor.SetMode(cursor.CursorBlink)
			m.editMailModel[m.inputSwitch].Cursor.Style = cursorStyle
		}
		m.editMailModel[m.inputSwitch], cmd = m.editMailModel[m.inputSwitch].Update(msg)
		return m, cmd
	}
	return m, tea.Quit
}

func (m ForwardList) View() string {
	v := lipgloss.NewStyle().MarginTop(1)
	var view string
	switch m.state {
	case inputState:
		var b strings.Builder
		for i := range m.editMailModel {
			b.WriteString(m.editMailModel[i].View())
			if i < len(m.editMailModel)-1 {
				b.WriteRune('\n')
			}
		}
		view = b.String()
	case forwardListState:
		view = m.tableModel.View()
	case errorState:
		view = strings.Join(strings.SplitAfterN(m.err.Error(), " ", m.windowWidth), "\n")
	}
	return v.Render(view)
}
