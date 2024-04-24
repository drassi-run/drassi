package secret_masker

// SecretMasker mask secret when print runner's log in gha console
// See https://github.com/actions/runner/blob/main/src/Sdk/DTLogging/Logging/ISecretMasker.cs
type SecretMasker interface {
	AddRegex(pattern string)
	AddValue(value string)
	Clone() SecretMasker
	MaskSecrets(input string) string
}
