package listener

import (
	"bytes"
	"crypto/cipher"
	"encoding/base64"

	"drassi.run/gha-runner/pkg/types"
)

// message provides a contract for receiving messages from the task orchestrator.
// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/TaskAgentMessage.cs
type message struct {
	// The message identifier
	Id int64 `json:"messageId,omitempty"`

	// The message type, describing the data contract found in Body
	Type string `json:"messageType,omitempty"`

	// The initialization vector used to encrypt this message
	IV []byte `json:"IV,omitempty"`

	// The body of the message. If the IV property is provided the body will need to be
	// decrypted using the Session.EncryptionKey value in addition to the IV.
	Body string `json:"body,omitempty"`
}

func (m *message) DecryptBody(key cipher.Block) ([]byte, error) {
	if len(m.IV) == 0 || key == nil {
		return []byte(m.Body), nil
	}

	cipherText, err := base64.StdEncoding.DecodeString(m.Body)
	if err != nil {
		return nil, err
	}
	plainText := make([]byte, len(cipherText))

	mode := cipher.NewCBCDecrypter(key, m.IV)
	mode.CryptBlocks(plainText, cipherText)

	plainText = unpad(plainText)
	plainText = bytes.TrimPrefix(plainText, types.Utf8BOM)
	return plainText, nil
}

// unpad removes PKCS7 padding from the data
func unpad(data []byte) []byte {
	length := len(data)
	unpadding := int(data[length-1])
	return data[:(length - unpadding)]
}
