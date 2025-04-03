package listener

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"hash"

	"drassi.run/gha-runner/pkg/messages"
	"drassi.run/gha-runner/pkg/types"
)

// Session represents a session for performing message exchanges from a runner (agent).
// https://github.com/actions/runner/blob/v2.323.0/src/Sdk/DTWebApi/WebApi/TaskAgentSession.cs
type Session struct {
	// The unique identifier for this session
	Id string `json:"sessionId,omitempty"` // UUID

	// The key used to encrypt message traffic for this session
	EncryptionKey *SessionKey `json:"encryptionKey,omitempty"`

	// The owner name of this session. Generally this will be the machine of origination
	OwnerName string `json:"ownerName,omitempty"`

	// The runner (agent) which is the target of the session
	Runner *types.RunnerReference `json:"agent,omitempty"`

	// whether to use FIPS compliant encryption scheme for job message key
	UseFipsEncryption bool `json:"useFipsEncryption,omitempty"`

	BrokerMigrationMessage *messages.BrokerMigration `json:"brokerMigrationMessage,omitempty"`
}

func (s *Session) GetKey(key *rsa.PrivateKey) (cipher.Block, error) {
	if s.EncryptionKey == nil || len(s.EncryptionKey.Value) == 0 {
		return nil, nil
	}
	eKey := s.EncryptionKey.Value
	if s.EncryptionKey.Encrypted {
		var hasher hash.Hash
		if s.UseFipsEncryption {
			hasher = crypto.SHA256.New()
		} else {
			hasher = crypto.SHA1.New()
		}

		if k, err := rsa.DecryptOAEP(hasher, rand.Reader, key, eKey, nil); err != nil || k == nil {
			return nil, err
		} else {
			eKey = k
		}
	}
	return aes.NewCipher(eKey)
}

// SessionKey represents a symmetric key used for message-level encryption for communication sent to an agent.
// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/TaskAgentSessionKey.cs
type SessionKey struct {
	// The value indicating whether the key value is encrypted. If this value is true, the Value property
	// should be decrypted using the RSA key exchanged with the server during registration.
	Encrypted bool `json:"encrypted,omitempty"`

	// The symmetric key value.
	Value []byte `json:"value,omitempty"`
}
