package main

import (
	"fmt"
	"log"
	"os"

	"github.com/antham/fnf/forward"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
	"golang.design/x/clipboard"
)

const envPrefix = "FNF"

type model struct {
	forward forward.Forward
	table   table.Model
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.updateList(msg)
}

func (m model) updateInput(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		}
	}
	return m, cmd
}

func (m model) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		case "c":
			clipboard.Write(clipboard.FmtText, []byte(m.table.SelectedRow()[1]))
		case "d":
			m.forward.Delete(m.table.SelectedRow()[0])
			m.table.SetRows(append(m.table.Rows()[:m.table.Cursor()], m.table.Rows()[m.table.Cursor()+1:]...))
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	v := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))
	return v.Render(m.table.View())
}

func main() {
	var endpoint, appKey, appSecret, consumerKey, domain string
	viper.SetEnvPrefix(envPrefix)
	viper.AutomaticEnv()
	for key, target := range map[string]*string{
		"OVH_ENDPOINT":     &endpoint,
		"OVH_APP_KEY":      &appKey,
		"OVH_APP_SECRET":   &appSecret,
		"OVH_CONSUMER_KEY": &consumerKey,
		"OVH_DOMAIN":       &domain,
	} {
		if !viper.IsSet(key) {
			log.Fatalf("%s_%s environment variable is not defined", envPrefix, key)
		}
		*target = viper.GetString(key)
	}
	forward, err := forward.NewOVHProvider(endpoint, appKey, appSecret, consumerKey, domain)
	if err != nil {
		log.Fatal(err)
	}
	m, err := newModel(forward)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func newModel(forward forward.Forward) (tea.Model, error) {
	columns := []table.Column{
		{Title: "ID", Width: 34},
		{Title: "From", Width: 40},
		{Title: "To", Width: 40},
	}
	rows := []table.Row{}
	list, err := forward.List()
	if err != nil {
		return nil, err
	}
	for _, l := range list {
		rows = append(rows, table.Row{l.ID, l.From, l.To})
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)
	return model{forward: forward, table: t}, nil
}
