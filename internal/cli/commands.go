package cli

import "strings"

type Command struct {
	Name        string
	Description string
}

var commands = []Command{
	{
		Name:        "register",
		Description: "create a new user",
	},
	{
		Name:        "login",
		Description: "login with username and password",
	},
	{
		Name:        "help",
		Description: "show available commands",
	},
	{
		Name:        "exit",
		Description: "quit the application",
	},
}

var authenticatedCommands = []Command{
	{
		Name:        "whoami",
		Description: "show current user details",
	},
	{
		Name:        "enable-2fa",
		Description: "enable TOTP-based MFA",
	},
	{
		Name:        "disable-2fa",
		Description: "disable MFA",
	},
	{
		Name:        "logout",
		Description: "end the current session",
	},
	{
		Name:        "help",
		Description: "show available commands",
	},
}

func findCommand(input string, authenticated bool) *Command {
	input = strings.TrimSpace(input)

	commandList := commands
	if authenticated {
		commandList = authenticatedCommands
	}

	for i := range commandList {
		if commandList[i].Name == input {
			return &commandList[i]
		}
	}

	return nil
}

func commandNames(authenticated bool) []string {
	commandList := commands
	if authenticated {
		commandList = authenticatedCommands
	}

	names := make([]string, 0, len(commandList))

	for _, command := range commandList {
		names = append(names, command.Name)
	}

	return names
}
