package commands

type Command interface {
	isCommand()
}

type WriteCommand interface {
	Command
}

type commandMarker struct{}

func (commandMarker) isCommand() {}
