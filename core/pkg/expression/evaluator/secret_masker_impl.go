package evaluator

import (
	"github.com/dungdm93/drassi/core/pkg/secret_masker"
)

// TODO next phase
// noOpSecretMasker is doing Noop currently
// Original implementations, see https://github.com/actions/runner/blob/main/src/Sdk/DTLogging/Logging/SecretMasker.cs
type noOpSecretMasker struct {
	originalValueSecrets map[string]struct{}
}

func (s noOpSecretMasker) AddRegex(pattern string) {
}

func (s noOpSecretMasker) AddValue(value string) {
}

func (s noOpSecretMasker) Clone() secret_masker.Interface {
	return noOpSecretMasker{originalValueSecrets: s.originalValueSecrets}
}

func (s noOpSecretMasker) MaskSecrets(input string) string {
	return input
}

func newNoOpSecretMasker() secret_masker.Interface {
	return &noOpSecretMasker{}
}
