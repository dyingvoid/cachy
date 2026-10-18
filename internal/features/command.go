package features

type Command interface {
	isCommand()
}

type ReplicationCommand interface {
	Command
}

type commandMarker struct{}

func (commandMarker) isCommand() {}
