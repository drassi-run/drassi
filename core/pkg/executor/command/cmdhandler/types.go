package cmdhandler

import "drassi.run/core/pkg/executor/runtime"

type SupportAddPath interface {
	AddPath(paths []string)
}

type SupportSetEnv interface {
	SetEnv(env map[string]string)
}

type SupportSetOutput interface {
	SetOutput(output map[string]string)
}

type SupportSaveState interface {
	SaveState(state map[string]string)
}

type SupportPathTranslator interface {
	PathTranslator() runtime.PathTranslator
}
