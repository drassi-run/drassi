package secret_masker

import (
	"github.com/dungdm93/drassi/core/pkg/executor/secret"
)

// SecretMasker mask value that match the secret regex
// See https://github.com/actions/runner/blob/main/src/Sdk/DTLogging/Logging/ISecretMasker.cs

type Interface interface {
	AddSecret(secret secret.Secret)
	Mask(input string) string
}
