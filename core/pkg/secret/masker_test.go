/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package secret

import (
	"errors"
	"io"
	"strings"
	"testing"

	mock_io "drassi.run/core/mock/sdk/io"
	xtypes "drassi.run/core/util/types"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type dummySecret struct {
	pos []*Position
}

func (d *dummySecret) In(input string) bool {
	return len(d.pos) > 0
}

func (d *dummySecret) At(input string) []*Position {
	return d.pos
}

func TestSecretMask(t *testing.T) {
	t.Run("no-match", func(tt *testing.T) {
		m := masker{}
		m.AddSecret(&dummySecret{})

		in := "0123456789"
		out := m.Mask(in)
		assert.Equal(tt, in, out)
	})

	t.Run("match-start", func(tt *testing.T) {
		m := masker{}
		m.AddSecret(&dummySecret{[]*Position{
			{Start: 0, End: 5},
		}})

		in := "0123456789"
		out := m.Mask(in)
		assert.Equal(tt, "***56789", out)
	})

	t.Run("match-end", func(tt *testing.T) {
		in := "0123456789"

		m := masker{}
		m.AddSecret(&dummySecret{[]*Position{
			{Start: 5, End: len(in)},
		}})

		out := m.Mask(in)
		assert.Equal(tt, "01234***", out)
	})

	t.Run("near", func(tt *testing.T) {
		in := "0123456789"

		m := masker{}
		m.AddSecret(&dummySecret{[]*Position{
			{Start: 2, End: 5},
			{Start: 5, End: 8},
		}})

		out := m.Mask(in)
		assert.Equal(tt, "01***89", out)
	})

	t.Run("overlap", func(tt *testing.T) {
		m := masker{}
		m.AddSecret(&dummySecret{[]*Position{
			{Start: 12, End: 20},
			{Start: 1, End: 4},
			{Start: 6, End: 9},
		}})

		in := "012345678901234567890123456789"
		out := m.Mask(in)
		assert.Equal(tt, "0***45***901***0123456789", out)

		m.AddSecret(&dummySecret{[]*Position{
			{Start: 2, End: 5},
		}})
		out = m.Mask(in)
		assert.Equal(tt, "0***5***901***0123456789", out)

		m.AddSecret(&dummySecret{[]*Position{
			{Start: 3, End: 8},
		}})
		out = m.Mask(in)
		assert.Equal(tt, "0***901***0123456789", out)
	})
}

func TestMaskReader(t *testing.T) {
	tests := map[string]xtypes.Pair[string, string]{
		"normal": {
			Key:   "first secret-token\nsecond password\nthird clean\n",
			Value: "first ***\nsecond ***\nthird clean\n",
		},
		"no-trailing-newline": {
			Key:   "secret-token",
			Value: "***",
		},
		"empty-line": {
			Key:   "first\n\nsecond password\n   \t\nthird clean\n",
			Value: "first\n\nsecond ***\n   \t\nthird clean\n",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			m := NewMasker()
			m.AddSecret(NewValueSecret("secret-token"))
			m.AddSecret(NewValueSecret("password"))

			in := strings.NewReader(tc.Key)
			out := MaskReader(m, in)
			scrubbed, err := io.ReadAll(out)
			assert.NoError(t, err)
			assert.Equal(t, tc.Value, string(scrubbed))
		})
	}
}

func TestMaskReader_ReadError(t *testing.T) {
	m := NewMasker()
	m.AddSecret(NewValueSecret("secret-token"))

	ctrl := gomock.NewController(t)
	in := mock_io.NewMockReader(ctrl)
	readErr := errors.New("read failed")
	gomock.InOrder(
		in.EXPECT().Read(gomock.Any()).
			DoAndReturn(func(p []byte) (int, error) {
				return copy(p, "The secret-token "), nil
			}),
		in.EXPECT().Read(gomock.Any()).
			DoAndReturn(func(p []byte) (int, error) {
				return copy(p, "will be masked"), readErr
			}),
	)

	r := MaskReader(m, in)

	out, err := io.ReadAll(r)
	assert.ErrorIs(t, err, readErr)
	assert.Equal(t, "The *** will be masked", string(out))
}
