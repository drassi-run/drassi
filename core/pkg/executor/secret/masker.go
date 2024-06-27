package secret

import (
	"slices"
	"strings"
)

type Masker interface {
	AddSecret(secret Secret)
	Mask(input string) string
}

func NewMasker() Masker {
	return &masker{}
}

type masker struct {
	secrets []Secret
}

func (m *masker) AddSecret(secret Secret) {
	m.secrets = append(m.secrets, secret)
}

func (m *masker) Mask(input string) string {
	pos := make([]*Position, 0)
	for _, secret := range m.secrets {
		pos = append(pos, secret.At(input)...)
	}
	if len(pos) == 0 {
		return input
	}

	// sort by Start ASC then End ASC
	slices.SortFunc(pos, func(a, b *Position) int {
		if a.Start != b.Start {
			return a.Start - b.Start
		}
		return a.End - b.End
	})

	mergePos := make([]*Position, 0)
	var curr *Position
	for _, p := range pos {
		if curr == nil || curr.End < p.Start {
			// clone SecretPosition
			curr = &Position{
				Start: p.Start,
				End:   p.End,
			}
			mergePos = append(mergePos, curr)
		} else {
			// overlap secrets
			curr.End = max(curr.End, p.End)
		}
	}

	builder := strings.Builder{}
	idx := 0
	for _, p := range mergePos {
		builder.WriteString(input[idx:p.Start])
		builder.WriteString("***")
		idx = p.End
	}
	builder.WriteString(input[idx:])

	return builder.String()
}
