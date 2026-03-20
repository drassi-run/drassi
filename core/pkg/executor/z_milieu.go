package executor

import (
	"drassi.run/core/pkg/executor/runtime"
	xtypes "drassi.run/core/util/types"
)

type Milieu interface {
	StepExecutor() StepExecutor

	AddPath(paths []string)
	SetEnv(env map[string]string)
	SetOutput(output map[string]string)
	SaveState(state map[string]string)

	// implementation may implement the following methods
	//CommandFile(cmd string) string
	//PathTranslator() runtime.PathTranslator
}

func NewMilieu(exec StepExecutor) Milieu {
	return &milieu{exec: exec}
}

type milieu struct {
	exec StepExecutor
}

func (s *milieu) StepExecutor() StepExecutor {
	return s.exec
}

func (s *milieu) AddPath(paths []string) {
	s.exec.JobExecutor().AddPath(paths)
}

func (s *milieu) SetEnv(env map[string]string) {
	for exec := s.exec; exec != nil; exec = exec.Parent() {
		exec.SetEnv(env)
	}
	s.exec.JobExecutor().SetEnv(env)
}

func (s *milieu) SetOutput(output map[string]string) {
	s.exec.SetOutput(output)
}

func (s *milieu) SaveState(state map[string]string) {
	s.exec.SaveState(state)
}

func (s *milieu) CommandFile(cmd string) string {
	return cmd + "_" + s.exec.StepSpec().Uid
}

func (s *milieu) PathTranslator() runtime.PathTranslator {
	for exec := s.exec.ActionExecutor(); exec != nil; {
		if pt, ok := s.exec.ActionExecutor().(interface{ PathTranslator() runtime.PathTranslator }); ok {
			return pt.PathTranslator()
		}
		if w, ok := exec.(xtypes.Unwrapper[ActionExecutor]); ok {
			exec = w.Unwrap()
			continue
		}
		break
	}
	return nil
}
