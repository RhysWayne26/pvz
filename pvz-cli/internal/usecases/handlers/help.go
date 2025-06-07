package handlers

import (
	"fmt"
)

// HandleHelp displays all available commands with their descriptions and usage.
func (f *DefaultFacadeHandler) HandleHelp() error {
	fmt.Println("Доступные команды:")
	for _, cmd := range AllCommands {
		fmt.Printf("  %-15s %s\n      Usage: %s\n", cmd.Name, cmd.Description, cmd.Usage)
	}
	return nil
}
