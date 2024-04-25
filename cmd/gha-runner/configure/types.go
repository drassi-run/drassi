package configure

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
	ID         int32           `json:"id"`
	Name       string          `json:"name"`
	Scope      string          `json:"scope"` // UUID
	PoolType   RunnerGroupType `json:"poolType"`
	Size       int32           `json:"size"`
	IsHosted   bool            `json:"isHosted"`
	IsInternal bool            `json:"isInternal"`
	IsLegacy   *bool           `json:"isLegacy"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/TaskAgentPool.cs
type RunnerGroup struct {
	RunnerGroupReference `json:",inline"`

	CreatedOn     time.Time `json:"createdOn"`
	AutoProvision *bool     `json:"autoProvision"`
	AutoSize      *bool     `json:"autoSize"`
	TargetSize    *int32    `json:"targetSize"`
	AgentCloudId  *int32    `json:"agentCloudId"`
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
