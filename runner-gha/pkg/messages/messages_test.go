/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package messages

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"testing"

	"drassi.run/gha-runner/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type MessageTestSuite struct {
	suite.Suite
	key    cipher.Block
	rawKey []byte
}

func TestMessageSuite(t *testing.T) {
	suite.Run(t, new(MessageTestSuite))
}

func (s *MessageTestSuite) SetupSuite() {
	s.rawKey = make([]byte, 32)
	_, err := rand.Read(s.rawKey)
	s.Require().NoError(err)

	s.key, err = aes.NewCipher(s.rawKey)
	s.Require().NoError(err)
}

func (s *MessageTestSuite) TestDecryptBody_NoEncryption() {
	t := s.T()
	body := "plain text body"

	// Case 1: IV is empty
	m := &Message{Body: body}
	decrypted, err := m.DecryptBody(s.key)
	assert.NoError(t, err)
	assert.Equal(t, []byte(body), decrypted)

	// Case 2: key is nil
	m = &Message{IV: []byte("some iv"), Body: body}
	decrypted, err = m.DecryptBody(nil)
	assert.NoError(t, err)
	assert.Equal(t, []byte(body), decrypted)
}

func (s *MessageTestSuite) TestDecryptBody_Success() {
	t := s.T()
	plainText := []byte("hello world")

	iv, cipherText := s.encrypt(plainText)
	m := &Message{
		IV:   iv,
		Body: base64.StdEncoding.EncodeToString(cipherText),
	}

	decrypted, err := m.DecryptBody(s.key)
	assert.NoError(t, err)
	assert.Equal(t, plainText, decrypted)
}

func (s *MessageTestSuite) TestDecryptBody_WithBOM() {
	t := s.T()
	plainText := append(types.Utf8BOM, []byte("hello with BOM")...)

	iv, cipherText := s.encrypt(plainText)
	m := &Message{
		IV:   iv,
		Body: base64.StdEncoding.EncodeToString(cipherText),
	}

	decrypted, err := m.DecryptBody(s.key)
	assert.NoError(t, err)
	assert.Equal(t, []byte("hello with BOM"), decrypted)
}

func (s *MessageTestSuite) TestDecryptBody_InvalidBase64() {
	t := s.T()
	m := &Message{
		IV:   []byte("1234567890123456"),
		Body: "invalid base64!!!",
	}

	_, err := m.DecryptBody(s.key)
	assert.Error(t, err)
}

func (s *MessageTestSuite) encrypt(plainText []byte) (iv, cipherText []byte) {
	// PKCS7 padding
	padLen := aes.BlockSize - (len(plainText) % aes.BlockSize)
	padded := append(plainText, make([]byte, padLen)...)
	for i := len(plainText); i < len(padded); i++ {
		padded[i] = byte(padLen)
	}

	iv = make([]byte, aes.BlockSize)
	_, err := rand.Read(iv)
	s.Require().NoError(err)

	cipherText = make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(s.key, iv)
	mode.CryptBlocks(cipherText, padded)
	return
}
