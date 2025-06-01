package handlers

// CommandHandler defines an interface for executing a CLI command.
type CommandHandler interface {
	Handle() error
}
