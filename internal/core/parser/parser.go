package core_parser

import (
	"fmt"
	"strings"

	"github.com/dyingvoid/cachy/internal/domain"
	domain_errors "github.com/dyingvoid/cachy/internal/domain/errors"
)

type Parser struct {
	commands map[domain.CommandName]int
}

func NewParser() *Parser {
	m := map[domain.CommandName]int{
		domain.CommandGet: 1,
		domain.CommandSet: 2,
	}

	return &Parser{
		commands: m,
	}
}

func (p *Parser) Parse(in string) (domain.Command, error) {
	fields := strings.Fields(in)
	if len(fields) == 0 {
		return domain.Command{}, domain_errors.ErrEmptyCommand
	}

	name := domain.CommandName(strings.ToUpper(fields[0]))
	expected, ok := p.commands[name]
	if !ok {
		return domain.Command{}, fmt.Errorf("%w: %s", domain_errors.ErrUnknownCommand, name)
	}

	args := fields[1:]
	if len(args) != expected {
		return domain.Command{}, fmt.Errorf(
			"%w: %s expects %d arg(s), got %d",
			domain_errors.ErrWrongArgCount, name, expected, len(args),
		)
	}

	return domain.Command{Name: name, Args: args}, nil
}
