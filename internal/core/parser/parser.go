package core_parser

import (
	"fmt"
	"strings"

	"github.com/dyingvoid/cachy/internal/domain"
)

// TODO: define errors separately

type Parser struct {
	commands map[string]int
}

func NewParser() *Parser {
	m := map[string]int{
		"GET": 1,
		"SET": 2,
	}

	return &Parser{
		commands: m,
	}
}

func (p *Parser) Parse(in string) (domain.Command, error) {
	fields := strings.Fields(in)
	if len(fields) == 0 {
		return domain.Command{}, fmt.Errorf("empty command")
	}

	name := strings.ToUpper(fields[0])
	expected, ok := p.commands[name]
	if !ok {
		return domain.Command{}, fmt.Errorf("parser: unknown command")
	}

	args := fields[1:]
	if len(args) != expected {
		return domain.Command{},
			fmt.Errorf("parser: %s expects %d arg(s), got %d", name, expected, len(args))
	}

	return domain.Command{Name: name, Args: args}, nil
}
