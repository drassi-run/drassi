/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package types

import (
	"time"

	"drassi.run/core/pkg/executor"
	"drassi.run/gha-runner/pkg/dotnet"
)

type RecordStore interface {
	RecordUid(stage executor.Stage, uid string) string
}

func NewList[T any](value []T) *List[T] {
	return &List[T]{
		Count: len(value),
		Value: value,
	}
}

type List[T any] struct {
	Count int `json:"count"`
	Value []T `json:"value"`
}

type GroupType string

const GroupTypeAutomation GroupType = "Automation"

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/TaskAgentPoolReference.cs#L23
type GroupReference struct {
	Id         int       `json:"id,omitempty"`
	Name       string    `json:"name,omitempty"`
	Scope      string    `json:"scope,omitempty"` // UUID
	GroupType  GroupType `json:"poolType,omitempty"`
	Size       int       `json:"size,omitempty"`
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
	TargetSize    *int      `json:"targetSize,omitempty"`
	AgentCloudId  *int      `json:"agentCloudId,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/TaskAgentStatus.cs
type RunnerStatus string

const (
	RunnerStatusOffline RunnerStatus = "offline"
	RunnerStatusOnline  RunnerStatus = "online"
	RunnerStatusBusy    RunnerStatus = "busy"
)

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/TaskAgentReference.cs#L17
type RunnerReference struct {
	Id                int          `json:"id,omitempty"`
	Name              string       `json:"name,omitempty"`
	Version           string       `json:"version,omitempty"`
	GroupId           int          `json:"runnerGroupId,omitempty"`
	GroupName         string       `json:"runnerGroupName,omitempty"`
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

	MaxParallelism  int        `json:"maxParallelism,omitempty"`
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

var Utf8BOM = []byte{'\xef', '\xbb', '\xbf'}
