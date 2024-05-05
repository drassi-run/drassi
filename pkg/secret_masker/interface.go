package secret_masker

// SecretMasker mask value that match the secret regex
// See https://github.com/actions/runner/blob/main/src/Sdk/DTLogging/Logging/ISecretMasker.cs

type Interface interface {
	AddRegex(pattern string)
	AddValue(value string)
	Clone() Interface
	MaskSecrets(input string) string
}
