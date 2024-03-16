package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"slices"

	"github.com/antham/fnf/forward"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
	"golang.design/x/clipboard"
	"golang.org/x/crypto/ssh/terminal"
)

const envPrefix = "FNF"

type model struct {
	forward      *forward.OVHProvider
	table        table.Model
	defaultEmail string
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
}

func (m model) View() string {
	v := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).MarginRight(1000)

	return v.Render(m.table.View())
}

func main() {
	var endpoint, appKey, appSecret, consumerKey, domain, defaultEmail string
	viper.SetEnvPrefix(envPrefix)
	viper.AutomaticEnv()
	for key, target := range map[string]*string{
		"OVH_ENDPOINT":     &endpoint,
		"OVH_APP_KEY":      &appKey,
		"OVH_APP_SECRET":   &appSecret,
		"OVH_CONSUMER_KEY": &consumerKey,
		"OVH_DOMAIN":       &domain,
		"DEFAULT_EMAIL":    &defaultEmail,
	} {
		if !viper.IsSet(key) {
			log.Fatalf("%s_%s environment variable is not defined", envPrefix, key)
		}
		*target = viper.GetString(key)
	}
	forward, err := forward.NewOVHProvider(endpoint, appKey, appSecret, consumerKey, domain, defaultEmail)
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

func newModel(forward *forward.OVHProvider) (tea.Model, error) {
	width, _, err := terminal.GetSize(0)
	if err != nil {
		return nil, err
	}
	columns := []table.Column{
		{Title: "Forward email", Width: width/2 - 3},
		{Title: "Destination", Width: width/2 - 3},
	}
	list, err := forward.List()
	if err != nil {
		log.Fatal(err)
	}

	rows := createRows(list)
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

func createRows(list []forward.ForwardInfo) []table.Row {
	rows := []table.Row{}
	for _, l := range list {
		rows = append(rows, table.Row{l.From, l.To})
	}
	return rows
}
