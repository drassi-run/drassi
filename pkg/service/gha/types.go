package gha

import (
	"crypto/rsa"
	"fmt"
	"math"
	"math/big"
	"time"
)

type RunnerGroupType string

const RunnerGroupTypeAutomation RunnerGroupType = "Automation"

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/TaskAgentPoolReference.cs#L23
type RunnerGroupReference struct {
	ID         int32           `json:"id,omitempty"`
	Name       string          `json:"name,omitempty"`
	Scope      string          `json:"scope,omitempty"` // UUID
	PoolType   RunnerGroupType `json:"poolType,omitempty"`
	Size       int32           `json:"size,omitempty"`
	IsHosted   bool            `json:"isHosted,omitempty"`
	IsInternal bool            `json:"isInternal,omitempty"`
	IsLegacy   *bool           `json:"isLegacy,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/TaskAgentPool.cs
type RunnerGroup struct {
	RunnerGroupReference `json:",inline"`

	CreatedOn     time.Time `json:"createdOn,omitempty"`
	AutoProvision *bool     `json:"autoProvision,omitempty"`
	AutoSize      *bool     `json:"autoSize,omitempty"`
	TargetSize    *int32    `json:"targetSize,omitempty"`
	AgentCloudId  *int32    `json:"agentCloudId,omitempty"`
}

type RunnerStatus string

const (
	RunnerStatusOffline RunnerStatus = "Offline"
	RunnerStatusOnline  RunnerStatus = "Online"
	RunnerStatusBusy    RunnerStatus = "Busy"
)

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/TaskAgentReference.cs#L17
type RunnerReference struct {
	Id                int32        `json:"id,omitempty"`
	Name              string       `json:"name,omitempty"`
	Version           string       `json:"version,omitempty"`
	Enabled           bool         `json:"enabled,omitempty"`
	Ephemeral         bool         `json:"ephemeral,omitempty"`
	Status            RunnerStatus `json:"status,omitempty"`
	OSDescription     string       `json:"OSDescription,omitempty"`
	ProvisioningState string       `json:"provisioningState,omitempty"`
	AccessPoint       string       `json:"accessPoint,omitempty"`
	DisableUpdate     bool         `json:"disableUpdate,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/TaskAgent.cs
type Runner struct {
	RunnerReference `json:",inline"`

	MaxParallelism  int32      `json:"maxParallelism,omitempty"`
	CreatedOn       *time.Time `json:"createdOn,omitempty"`
	StatusChangedOn *time.Time `json:"statusChangedOn,omitempty"`

	// The request which is currently assigned to this runner
	// AssignedRequest RunnerJobRequest

	// The last request which was completed by this runner
	// LastCompletedRequest RunnerJobRequest

	Labels        []RunnerLabel       `json:"labels,omitempty"`
	Authorization RunnerAuthorization `json:"authorization,omitempty"`
}

type RunnerLabel struct {
	Id   int             `json:"id,omitempty"`
	Name string          `json:"name,omitempty"`
	Type RunnerLabelType `json:"type,omitempty"`
}

type RunnerLabelType string

const (
	RunnerLabelTypeSystem RunnerLabelType = "system"
	RunnerLabelTypeUser   RunnerLabelType = "user"
)

type RunnerAuthorization struct {
	AuthorizationUrl string           `json:"authorizationUrl,omitempty"`
	ClientId         string           `json:"clientId,omitempty"`
	PublicKey        *RunnerPublicKey `json:"publicKey,omitempty"`
}

type RunnerPublicKey struct {
	Exponent []byte `json:"exponent,omitempty"`
	Modulus  []byte `json:"modulus,omitempty"`
}

func NewRunnerPublicKey(pubkey *rsa.PublicKey) *RunnerPublicKey {
	bigE := big.NewInt(int64(pubkey.E))
	return &RunnerPublicKey{
		Exponent: bigE.Bytes(),
		Modulus:  pubkey.N.Bytes(),
	}
}

func (pk *RunnerPublicKey) ToRsaPublicKey() (*rsa.PublicKey, error) {
	mod := new(big.Int).SetBytes(pk.Modulus)
	exp := new(big.Int).SetBytes(pk.Exponent)

	var e int64
	if !exp.IsInt64() {
		return nil, fmt.Errorf("%s can be represented as an int64", exp)
	} else {
		e = exp.Int64()
		if e > math.MaxInt {
			return nil, fmt.Errorf("%d integer overflow", e)
		}
		if e <= 0 {
			return nil, fmt.Errorf("%d must be positive number", e)
		}
	}

	pubkey := rsa.PublicKey{
		N: mod,
		E: int(e),
	}
	return &pubkey, nil
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/TaskAgentSession.cs
type RunnerSession struct {
	// The unique identifier for this session
	Id string `json:"sessionId,omitempty"` // UUID

	// The key used to encrypt message traffic for this session
	EncryptionKey *RunnerSessionKey `json:"encryptionKey,omitempty"`

	// The owner name of this session. Generally this will be the machine of origination
	OwnerName string `json:"ownerName,omitempty"`

	// The runner (agent) which is the target of the session
	Runner *RunnerReference `json:"agent,omitempty"`

	// whether to use FIPS compliant encryption scheme for job message key
	UseFipsEncryption bool `json:"useFipsEncryption,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/TaskAgentSessionKey.cs
type RunnerSessionKey struct {
	// The value indicating whether the key value is encrypted. If this value is true, the Value property
	// should be decrypted using the RSA key exchanged with the server during registration.
	Encrypted bool `json:"encrypted,omitempty"`

	// The symmetric key value.
	Value []byte `json:"value,omitempty"`
}

type Message struct {
	// The message identifier
	Id string `json:"messageId,omitempty"`

	// The message type, describing the data contract found in Body
	Type string `json:"messageType,omitempty"`

	// The initialization vector used to encrypt this message
	IV []byte `json:"IV,omitempty"`

	// The body of the message. If the IV property is provided the body will need to be
	// decrypted using the Session.EncryptionKey value in addition to the IV.
	Body string `json:"body,omitempty"`
}
