package secret_masker

// ISecretMasker mask secret when print runner's log in gha console
// See https://github.com/actions/runner/blob/main/src/Sdk/DTLogging/Logging/ISecretMasker.cs
type ISecretMasker interface {
	AddRegex(pattern string)
	AddValue(value string)
	Clone() ISecretMasker
	MaskSecrets(input string) string
}
