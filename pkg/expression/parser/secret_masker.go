package parser

import (
	"github.com/dungdm93/drasi/pkg/expression/interfaces"
)

// TODO next phase

// SecretMasker is doing Noop currently
type SecretMasker struct {
	originalValueSecrets map[string]struct{}
}

func (s SecretMasker) AddRegex(pattern string) {
}

func (s SecretMasker) AddValue(value string) {
}

func (s SecretMasker) Clone() interfaces.ISecretMasker {
	// TODO implement me
	return s
}

func (s SecretMasker) MaskSecrets(input string) string {
	return input
}

func NewSecretMasker() interfaces.ISecretMasker {
	return &SecretMasker{}
}
