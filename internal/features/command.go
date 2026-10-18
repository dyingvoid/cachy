package features

type Command interface {
	isCommand()
}

type commandMarker struct{}

func (commandMarker) isCommand() {}
