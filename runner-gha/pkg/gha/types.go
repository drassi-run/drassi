/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package gha

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"hash"
	"time"

	"drassi.run/gha-runner/pkg/dotnet"
)

type GroupType string

const GroupTypeAutomation GroupType = "Automation"

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/TaskAgentPoolReference.cs#L23
type GroupReference struct {
	ID         int32     `json:"id,omitempty"`
	Name       string    `json:"name,omitempty"`
	Scope      string    `json:"scope,omitempty"` // UUID
	GroupType  GroupType `json:"poolType,omitempty"`
	Size       int32     `json:"size,omitempty"`
	IsHosted   bool      `json:"isHosted,omitempty"`
	IsInternal bool      `json:"isInternal,omitempty"`
	IsLegacy   *bool     `json:"isLegacy,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/TaskAgentPool.cs
type Group struct {
	GroupReference `json:",inline"`

	CreatedOn     time.Time `json:"createdOn,omitempty"`
	AutoProvision *bool     `json:"autoProvision,omitempty"`
	AutoSize      *bool     `json:"autoSize,omitempty"`
	TargetSize    *int32    `json:"targetSize,omitempty"`
	AgentCloudId  *int32    `json:"agentCloudId,omitempty"`
}

type RunnerStatus string

const (
	RunnerStatusOffline RunnerStatus = "offline"
	RunnerStatusOnline  RunnerStatus = "online"
	RunnerStatusBusy    RunnerStatus = "busy"
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

	Labels        []Label       `json:"labels,omitempty"`
	Authorization Authorization `json:"authorization,omitempty"`
}

type Label struct {
	Id   int       `json:"id,omitempty"`
	Name string    `json:"name,omitempty"`
	Type LabelType `json:"type,omitempty"`
}

type LabelType string

const (
	LabelTypeSystem LabelType = "system"
	LabelTypeUser   LabelType = "user"
)

type Authorization struct {
	AuthorizationUrl string            `json:"authorizationUrl,omitempty"`
	ClientId         string            `json:"clientId,omitempty"`
	PublicKey        *dotnet.PublicKey `json:"publicKey,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/TaskAgentSession.cs
type Session struct {
	// The unique identifier for this session
	Id string `json:"sessionId,omitempty"` // UUID

	// The key used to encrypt message traffic for this session
	EncryptionKey *SessionKey `json:"encryptionKey,omitempty"`

	// The owner name of this session. Generally this will be the machine of origination
	OwnerName string `json:"ownerName,omitempty"`

	// The runner (agent) which is the target of the session
	Runner *RunnerReference `json:"agent,omitempty"`

	// whether to use FIPS compliant encryption scheme for job message key
	UseFipsEncryption bool `json:"useFipsEncryption,omitempty"`
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

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/TaskAgentSessionKey.cs
type SessionKey struct {
	// The value indicating whether the key value is encrypted. If this value is true, the Value property
	// should be decrypted using the RSA key exchanged with the server during registration.
	Encrypted bool `json:"encrypted,omitempty"`

	// The symmetric key value.
	Value []byte `json:"value,omitempty"`
}
