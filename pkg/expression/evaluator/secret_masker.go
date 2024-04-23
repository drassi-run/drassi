package evaluator

import (
	"github.com/dungdm93/drasi/pkg/expression"
)

// TODO next phase

// secretMasker is doing Noop currently
type secretMasker struct {
	originalValueSecrets map[string]struct{}
}

func (s secretMasker) AddRegex(pattern string) {
}

func (s secretMasker) AddValue(value string) {
}

func (s secretMasker) Clone() expression.ISecretMasker {
	return secretMasker{originalValueSecrets: s.originalValueSecrets}
}

func (s secretMasker) MaskSecrets(input string) string {
	return input
}

func newSecretMasker() expression.ISecretMasker {
	return &secretMasker{}
}
