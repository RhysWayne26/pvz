package handlers

import (
	"fmt"
	"pvz-cli/internal/usecases/cli"
)

// HelpCommandHandler handles the help command.
type HelpCommandHandler struct{}

// NewHelpCommandHandler creates a new instance of HelpCommandHandler.
func NewHelpCommandHandler() *HelpCommandHandler {
	return &HelpCommandHandler{}
}

// Handle displays all available commands with their descriptions and usage.
func (h *HelpCommandHandler) Handle() error {
	fmt.Println("Доступные команды:")
	for _, cmd := range cli.AllCommands {
		fmt.Printf("  %-15s %s\n      Usage: %s\n", cmd.Name, cmd.Description, cmd.Usage)
	}
	return nil
}
