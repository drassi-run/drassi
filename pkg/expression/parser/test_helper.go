package parser

import (
	"github.com/dungdm93/drasi/pkg/expression/interfaces"
)

type MockGithubContext struct {
	Actor string
}

func NewMockGithubContext(actor string) *MockGithubContext {
	return &MockGithubContext{Actor: actor}
}

func (g MockGithubContext) Count() int {
	return 1
}

func (g MockGithubContext) Keys() []string {
	return []string{"actor"}
}

func (g MockGithubContext) Values() []any {
	return []any{g.Actor}
}

func (g MockGithubContext) ContainsKey(key string) (exist bool) {
	return key == "actor"
}

func (g MockGithubContext) GetValue(key string) (exist bool, value any) {
	if key == "actor" {
		return true, g.Actor
	}
	return false, nil
}

func (g MockGithubContext) Enumerator() *interfaces.Enumerator {
	return interfaces.NewEnumerator(g.Actor)
}
