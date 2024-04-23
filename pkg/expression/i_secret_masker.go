package expression

type ISecretMasker interface {
	AddRegex(pattern string)
	AddValue(value string)
	Clone() ISecretMasker
	MaskSecrets(input string) string
}
