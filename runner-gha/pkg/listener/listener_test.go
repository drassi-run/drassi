/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package listener

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"testing"

	"drassi.run/gha-runner/pkg/messages"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ListenerTestSuite struct {
	suite.Suite
	key []byte
	ss  *Session
}

func TestListenerSuite(t *testing.T) {
	suite.Run(t, new(ListenerTestSuite))
}

func (s *ListenerTestSuite) SetupSuite() {
	s.key = randBytes(32)
	s.ss = &Session{
		EncryptionKey: &SessionKey{
			Value: s.key,
		},
	}
}

func (s *ListenerTestSuite) TestSetSession() {
	t := s.T()
	l := new(baseListener)

	err := l.SetSession(s.ss)
	assert.NoError(t, err)
	assert.Equal(t, s.ss, l.Session())
	assert.NotNil(t, l.encKey)
}

func (s *ListenerTestSuite) TestDecryptMessage() {
	t := s.T()

	l := new(baseListener)
	err := l.SetSession(s.ss)
	require.NoError(t, err)

	plainText := []byte("hello world")

	// PKCS7 padding
	padLen := aes.BlockSize - (len(plainText) % aes.BlockSize)
	padded := append(plainText, make([]byte, padLen)...)
	for i := len(plainText); i < len(padded); i++ {
		padded[i] = byte(padLen)
	}

	iv := randBytes(16)
	cipherText := make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(l.encKey, iv)
	mode.CryptBlocks(cipherText, padded)

	m := &messages.Message{
		Id:   123,
		Type: "test",
		IV:   iv,
		Body: base64.StdEncoding.EncodeToString(cipherText),
	}

	decrypted, err := l.DecryptMessage(m)
	assert.NoError(t, err)
	assert.Equal(t, m.Id, decrypted.Id)
	assert.Equal(t, m.Type, decrypted.Type)
	assert.Equal(t, plainText, decrypted.Body)
}
