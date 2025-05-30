package handlers

import (
	"fmt"
	"pvz-cli/internal/usecases/cli"
)

// HandleHelpCommand displays all available commands with their descriptions and usage
func HandleHelpCommand() {
	fmt.Println("Доступные команды:")
	for _, cmd := range cli.AllCommands {
		fmt.Printf("  %-15s %s\n      Usage: %s\n", cmd.Name, cmd.Description, cmd.Usage)
	}
}
