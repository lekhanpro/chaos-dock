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
		"Chaos-Dock\n\n" +
			"Available workflows:\n" +
			"- Init config: go run ./cmd/chaos-dock -init-config -config chaos.yaml\n" +
			"- Validate config: go run ./cmd/chaos-dock -validate-config -config chaos.yaml\n" +
			"- List containers: go run ./cmd/chaos-dock -list\n" +
			"- TUI mode: go run ./cmd/chaos-dock\n" +
			"- Run once: go run ./cmd/chaos-dock -run-once -config chaos.yaml\n" +
			"- Run scheduled: go run ./cmd/chaos-dock -run-scheduled -config chaos.yaml\n" +
			"- Panic rollback: go run ./cmd/chaos-dock -panic -targets \"postgres,redis\"\n\n" +
			"Press q to quit.\n",
	)
}
