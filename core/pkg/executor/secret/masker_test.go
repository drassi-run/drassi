package secret

import (
	"gotest.tools/v3/assert"
	"testing"
)

type dummySecret struct {
	pos []*Position
}

func (d *dummySecret) At(input string) []*Position {
	return d.pos
}

func TestSecretMask(t *testing.T) {
	t.Run("no-match", func(tt *testing.T) {
		masker := Masker{}
		masker.AddSecret(&dummySecret{})

		in := "0123456789"
		out := masker.Mask(in)
		assert.Equal(tt, in, out)
	})

	t.Run("match-start", func(tt *testing.T) {
		masker := Masker{}
		masker.AddSecret(&dummySecret{[]*Position{
			{Start: 0, End: 5},
		}})

		in := "0123456789"
		out := masker.Mask(in)
		assert.Equal(tt, "***56789", out)
	})

	t.Run("match-end", func(tt *testing.T) {
		in := "0123456789"

		masker := Masker{}
		masker.AddSecret(&dummySecret{[]*Position{
			{Start: 5, End: len(in)},
		}})

		out := masker.Mask(in)
		assert.Equal(tt, "01234***", out)
	})

	t.Run("near", func(tt *testing.T) {
		in := "0123456789"

		masker := Masker{}
		masker.AddSecret(&dummySecret{[]*Position{
			{Start: 2, End: 5},
			{Start: 5, End: 8},
		}})

		out := masker.Mask(in)
		assert.Equal(tt, "01***89", out)
	})

	t.Run("overlap", func(tt *testing.T) {
		masker := Masker{}
		masker.AddSecret(&dummySecret{[]*Position{
			{Start: 12, End: 20},
			{Start: 1, End: 4},
			{Start: 6, End: 9},
		}})

		in := "012345678901234567890123456789"
		out := masker.Mask(in)
		assert.Equal(tt, "0***45***901***0123456789", out)

		masker.AddSecret(&dummySecret{[]*Position{
			{Start: 2, End: 5},
		}})
		out = masker.Mask(in)
		assert.Equal(tt, "0***5***901***0123456789", out)

		masker.AddSecret(&dummySecret{[]*Position{
			{Start: 3, End: 8},
		}})
		out = masker.Mask(in)
		assert.Equal(tt, "0***901***0123456789", out)
	})
}
