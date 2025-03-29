package listener

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"hash"

	"drassi.run/gha-runner/pkg/message"
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

	BrokerMigrationMessage *message.BrokerMigration `json:"brokerMigrationMessage,omitempty"`
}

func (s *Session) GetEncryptionKey(key *rsa.PrivateKey) ([]byte, error) {
	if s.EncryptionKey == nil || len(s.EncryptionKey.Value) == 0 {
		return nil, nil
	}
	if s.EncryptionKey.Encrypted {
		var hasher hash.Hash
		if s.UseFipsEncryption {
			hasher = crypto.SHA256.New()
		} else {
			hasher = crypto.SHA1.New()
		}

		return rsa.DecryptOAEP(hasher, rand.Reader, key, s.EncryptionKey.Value, nil)
	}
	return s.EncryptionKey.Value, nil
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
