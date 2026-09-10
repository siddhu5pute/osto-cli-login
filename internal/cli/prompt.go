package cli

import (
	"fmt"
	"strings"

	"github.com/chzyer/readline"
)

type Prompt struct {
	reader *readline.Instance
}

func NewPrompt() (*Prompt, error) {
	reader, err := readline.NewEx(&readline.Config{
		Prompt:          "osto> ",
		HistoryFile:     ".osto_history",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
		AutoComplete: readline.NewPrefixCompleter(
			readline.PcItem("register"),
			readline.PcItem("login"),
			readline.PcItem("whoami"),
			readline.PcItem("enable-2fa"),
			readline.PcItem("disable-2fa"),
			readline.PcItem("logout"),
			readline.PcItem("help"),
			readline.PcItem("exit"),
		),
	})
	if err != nil {
		return nil, fmt.Errorf("creating prompt: %w", err)
	}

	return &Prompt{
		reader: reader,
	}, nil
}

func (p *Prompt) ReadLine() (string, error) {
	line, err := p.reader.Readline()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(line), nil
}

func (p *Prompt) ReadPassword(prompt string) (string, error) {
	password, err := p.reader.ReadPassword(prompt)
	if err != nil {
		return "", fmt.Errorf("reading password: %w", err)
	}

	return strings.TrimSpace(string(password)), nil
}

func (p *Prompt) Close() error {
	return p.reader.Close()
}
