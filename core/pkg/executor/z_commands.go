package executor

type SupportCommands interface {
	AddPath(paths []string)
	SetEnv(env map[string]string)
	SetOutput(output map[string]string)
	SaveState(state map[string]string)

	// implementation may implement the following methods
	//CommandFile(cmd string) string
	//PathTranslator() runtime.PathTranslator
}

func NewSupportCommands(exec StepExecutor) SupportCommands {
	panic("TODO implement me")
}
