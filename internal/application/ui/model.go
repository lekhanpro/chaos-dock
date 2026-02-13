package ui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	ctx context.Context
}

func NewModel(ctx context.Context) Model {
	return Model{ctx: ctx}
}

func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		<-m.ctx.Done()
		return tea.QuitMsg{}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m Model) View() string {
	return fmt.Sprintf(
		"Chaos-Dock (Phase 1)\n\n"+
			"- Docker discovery and experiment runner wire-up in progress.\n"+
			"- Press q to quit.\n",
	)
}
