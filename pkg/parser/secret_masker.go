package parser

// TODO next phase

type ISecretMasker interface {
	AddRegex(pattern string)
	AddValue(value string)
	AddValueEncoder(encoder *ValueEncoder)
	Clone() ISecretMasker
	MaskSecrets(input string) string
}

type ValueEncoder struct {
}

// SecretMasker is doing Noop currently
type SecretMasker struct {
	originalValueSecrets map[string]struct{}
}

func (s SecretMasker) AddRegex(pattern string) {
}

func (s SecretMasker) AddValue(value string) {
}

func (s SecretMasker) AddValueEncoder(encoder *ValueEncoder) {
	// TODO implement me
}

func (s SecretMasker) Clone() ISecretMasker {
	// TODO implement me
	return s
}

func (s SecretMasker) MaskSecrets(input string) string {
	return input
}

func NewSecretMasker() ISecretMasker {
	return &SecretMasker{}
}
