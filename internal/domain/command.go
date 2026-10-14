package domain

type CommandName string

const (
	CommandGet CommandName = "GET"
	CommandSet CommandName = "SET"
)

func (c CommandName) String() string {
	return string(c)
}

type Command struct {
	Name CommandName
	Args []string
}
