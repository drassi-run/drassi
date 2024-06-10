package evaluator

import (
	"github.com/dungdm93/drassi/core/pkg/executor/secret"
)

type noOpSecretMasker struct {}

func (n *noOpSecretMasker) AddSecret(secret secret.Secret) {
}

func (n *noOpSecretMasker) Mask(input string) string {
	return input
}
