/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package listener

import (
	"crypto/aes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"hash"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type SessionTestSuite struct {
	suite.Suite
	privKey *rsa.PrivateKey
}

func TestSessionSuite(t *testing.T) {
	suite.Run(t, new(SessionTestSuite))
}

func (s *SessionTestSuite) SetupSuite() {
	var err error
	s.privKey, err = rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(s.T(), err)
}

func (s *SessionTestSuite) TestGetKey_NoKey() {
	t := s.T()
	ss := &Session{}

	encKey, err := ss.GetKey(s.privKey)
	assert.NoError(t, err)
	assert.Nil(t, encKey)
}

func (s *SessionTestSuite) TestGetKey_Plain() {
	t := s.T()
	key := randBytes(32)

	ss := &Session{
		EncryptionKey: &SessionKey{
			Encrypted: false,
			Value:     key,
		},
	}

	block, err := ss.GetKey(s.privKey)
	assert.NoError(t, err)
	assert.NotNil(t, block)

	// Verify it's the right key by encrypting/decrypting
	data := randBytes(aes.BlockSize)
	encrypted := make([]byte, aes.BlockSize)
	decrypted := make([]byte, aes.BlockSize)
	block.Encrypt(encrypted, data)
	block.Decrypt(decrypted, encrypted)
	assert.Equal(t, data, decrypted)
}

func (s *SessionTestSuite) TestGetKey_Encrypted_SHA1() {
	s.test_GetKey_Encrypted(sha1.New(), false)
}

func (s *SessionTestSuite) TestGetKey_Encrypted_SHA256() {
	s.test_GetKey_Encrypted(sha256.New(), true)
}

//goland:noinspection GoSnakeCaseUsage
func (s *SessionTestSuite) test_GetKey_Encrypted(h hash.Hash, fips bool) {
	t := s.T()
	key := randBytes(32)

	encKey, err := rsa.EncryptOAEP(h, rand.Reader, &s.privKey.PublicKey, key, nil)
	require.NoError(t, err)

	ss := &Session{
		EncryptionKey: &SessionKey{
			Encrypted: true,
			Value:     encKey,
		},
		UseFipsEncryption: fips,
	}

	block, err := ss.GetKey(s.privKey)
	assert.NoError(t, err)
	assert.NotNil(t, block)
}

func randBytes(n int) []byte {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return b
}
