/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package message

import (
	"time"
)

const (
	TypeAgentRefresh            = "AgentRefresh"
	TypeRunnerRefresh           = "RunnerRefresh"
	TypeRunnerShutdown          = "RunnerShutdown"
	TypeJobCancelMessage        = "JobCancellation"
	TypeRunnerJobRequest        = "RunnerJobRequest"
	TypePipelineAgentJobRequest = "PipelineAgentJobRequest"
	TypeForceTokenRefresh       = "ForceTokenRefresh"
)

type Duration struct {
	time.Duration
}

type Time struct {
	time.Time
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/AgentRefreshMessage.cs
type AgentRefresh struct {
	AgentId       int32    `actions:"agentId,omitempty"`
	Timeout       Duration `actions:"timeout,omitempty"`
	TargetVersion string   `actions:"targetVersion,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/RunnerRefreshMessage.cs
type RunnerRefresh struct {
	TargetVersion  string `actions:"target_version,omitempty"`
	DownloadUrl    string `actions:"download_url,omitempty"`
	SHA256Checksum string `actions:"sha256_checksum,omitempty"`
	OS             string `actions:"os,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/RunnerShutdownMessage.cs
type RunnerShutdown struct {
	Reason string `actions:"reason,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/JobCancelMessage.cs
type JobCancel struct {
	JobId   string   `actions:"jobId,omitempty"`
	Timeout Duration `actions:"timeout,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Listener/RunnerJobRequestRef.cs
type RunnerJobRequest struct {
	Id              string `actions:"id,omitempty"`
	RunnerRequestId string `actions:"runner_request_id,omitempty"`
	RunServiceUrl   string `actions:"run_service_url,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/AgentJobRequestMessage.cs
type PipelineAgentJobRequest struct {
	MessageType          string                   `actions:"messageType,omitempty"`
	RequestId            int64                    `actions:"requestId,omitempty"`
	Plan                 PlanReference            `actions:"plan,omitempty"`
	Timeline             TimelineReference        `actions:"timeline,omitempty"`
	JobId                string                   `actions:"jobId,omitempty"`
	JobName              string                   `actions:"jobName,omitempty"`
	JobDisplayName       string                   `actions:"jobDisplayName,omitempty"`
	JobContainer         *TemplateToken           `actions:"jobContainer,omitempty"`
	JobServiceContainers *TemplateToken           `actions:"jobServiceContainers,omitempty"`
	JobOutputs           *TemplateToken           `actions:"jobOutputs,omitempty"`
	LockedUntil          Time                     `actions:"lockedUntil,omitempty"`
	Resources            *JobResources            `actions:"resources,omitempty"`
	ContextData          ContextData              `actions:"contextData,omitempty"`
	Workspace            *WorkspaceOptions        `actions:"workspace,omitempty"`
	MaskHints            []MaskHint               `actions:"mask,omitempty"`
	Env                  []TemplateToken          `actions:"environmentVariables,omitempty"`
	Defaults             []TemplateToken          `actions:"defaults,omitempty"`
	Environment          *EnvironmentReference    `actions:"actionsEnvironment,omitempty"`
	Snapshot             *TemplateToken           `actions:"snapshot,omitempty"`
	Variables            map[string]VariableValue `actions:"variables,omitempty"`
	Steps                []JobStep                `actions:"steps,omitempty"`
	FileTable            []string                 `actions:"fileTable,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/TaskOrchestrationPlanReference.cs
type PlanReference struct {
	ScopeIdentifier  string `actions:"scopeIdentifier,omitempty"` // UUID
	PlanType         string `actions:"planType,omitempty"`
	Version          int32  `actions:"version,omitempty"`
	PlanId           string `actions:"planId,omitempty"` // UUID
	PlanGroup        string `actions:"planGroup,omitempty"`
	ArtifactUri      string `actions:"artifactUri,omitempty"`      // URI
	ArtifactLocation string `actions:"artifactLocation,omitempty"` // URI
	ContainerId      int64  `actions:"containerId,omitempty"`
	Definition       Owner  `actions:"definition,omitempty"`
	Owner            Owner  `actions:"owner,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/TaskOrchestrationOwner.cs
type Owner struct {
	Id    int32          `actions:"id,omitempty"`
	Name  string         `actions:"name,omitempty"`
	Links map[string]any `actions:"_links,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/TimelineReference.cs
type TimelineReference struct {
	Id       string `actions:"id,omitempty"` // UUID
	ChangeId int32  `actions:"changeId,omitempty"`
	Location string `actions:"location,omitempty"` // URI
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTObjectTemplating/ObjectTemplating/Tokens/TemplateToken.cs
type TemplateToken struct {
	Type   TokenType `actions:"type,omitempty"`
	File   int32     `actions:"file,omitempty"`
	Line   int32     `actions:"line,omitempty"`
	Column int32     `actions:"column,omitempty"`

	String    string                                   `actions:"lit,omitempty"`       // StringToken (type=0)
	Number    float64                                  `actions:"num,omitempty"`       // NumberToken (type=6)
	Boolean   bool                                     `actions:"bool,omitempty"`      // BooleanToken (type=5)
	Directive string                                   `actions:"directive,omitempty"` // InsertExpressionToken (type=4)
	Expr      string                                   `actions:"expr,omitempty"`      // BasicExpressionToken (type=3)
	Seq       []*TemplateToken                         `actions:"seq,omitempty"`       // SequenceToken (type=1)
	Map       []KVPair[*TemplateToken, *TemplateToken] `actions:"map,omitempty"`       // MappingToken (type=2)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTObjectTemplating/ObjectTemplating/Tokens/TokenType.cs
type TokenType int32

const (
	TokenTypeString           TokenType = 0
	TokenTypeSequence         TokenType = 1
	TokenTypeMapping          TokenType = 2
	TokenTypeBasicExpression  TokenType = 3
	TokenTypeInsertExpression TokenType = 4
	TokenTypeNumber           TokenType = 5
	TokenTypeBoolean          TokenType = 6
	TokenTypeNull             TokenType = 7
)

type KVPair[K, V any] struct {
	Key   K `actions:"key,omitempty"`
	Value V `actions:"value,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/JobResources.cs
type JobResources struct {
	Endpoints    []ServiceEndpoint    `actions:"endpoints,omitempty"`
	Containers   []ContainerResource  `actions:"containers,omitempty"`
	Repositories []RepositoryResource `actions:"repositories,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/ServiceEndpointLegacy/ServiceEndpoint.cs
type ServiceEndpoint struct {
	Id              string                `actions:"id,omitempty"` // UUID
	Name            string                `actions:"name,omitempty"`
	Type            string                `actions:"type,omitempty"`
	Owner           string                `actions:"owner,omitempty"`
	Url             string                `actions:"url,omitempty"` // URI
	Description     string                `actions:"description,omitempty"`
	Authorization   EndpointAuthorization `actions:"authorization,omitempty"`
	GroupScopeId    string                `actions:"groupScopeId,omitempty"` // UUID
	Data            map[string]string     `actions:"data,omitempty"`
	IsShared        bool                  `actions:"isShared,omitempty"`
	IsReady         bool                  `actions:"isReady,omitempty"`
	OperationStatus any                   `actions:"operationStatus,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/ServiceEndpointLegacy/EndpointAuthorization.cs
type EndpointAuthorization struct {
	Scheme     string            `actions:"scheme,omitempty"`
	Parameters map[string]string `actions:"parameters,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/Resource.cs
type Resource struct {
	Alias      string                   `actions:"alias,omitempty"`
	Endpoint   ServiceEndpointReference `actions:"endpoint,omitempty"`
	Properties ResourceProperties       `actions:"properties,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/ServiceEndpointReference.cs
type ServiceEndpointReference struct {
	Id   string `actions:"id,omitempty"`   // UUID
	Name string `actions:"name,omitempty"` // ExpressionValue<String>
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/ResourceProperties.cs
type ResourceProperties struct {
	Count int32          `actions:"count,omitempty"`
	Items map[string]any `actions:"items,omitempty"` // IDictionary<String, JToken>
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/ContainerResource.cs
type ContainerResource struct {
	Resource `actions:",inline"`

	Image   string            `actions:"image,omitempty"`
	Env     map[string]string `actions:"env,omitempty"`
	Ports   []string          `actions:"ports,omitempty"`
	Volumes []string          `actions:"volumes,omitempty"`
	Options string            `actions:"options,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/RepositoryResource.cs
type RepositoryResource struct {
	Resource `actions:",inline"`

	Id      string `actions:"id,omitempty"`
	Type    string `actions:"type,omitempty"`
	Url     string `actions:"url,omitempty"` // URI
	Version string `actions:"version,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/WorkspaceOptions.cs
type WorkspaceOptions struct {
	Clean string `actions:"clean,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/MaskHint.cs
type MaskHint struct {
	Type  MaskType `actions:"type,omitempty"`
	Value string   `actions:"value,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/MaskType.cs
type MaskType string

const (
	MaskTypeRegex    MaskType = "regex"
	MaskTypeVariable MaskType = "variable"
)

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/ActionsEnvironmentReference.cs
type EnvironmentReference struct {
	Name string        `actions:"name,omitempty"`
	Url  TemplateToken `actions:"url,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/VariableValue.cs
type VariableValue struct {
	Value    string `actions:"value,omitempty"`
	IsSecret bool   `actions:"isSecret,omitempty"`
}

type ContextData map[string]any

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/JobStep.cs
type JobStep struct {
	// Step
	Id          string `actions:"id,omitempty"`
	Type        string `actions:"type,omitempty"`
	Name        string `actions:"name,omitempty"`
	DisplayName string `actions:"displayName,omitempty"`
	Enabled     bool   `actions:"enabled,omitempty"`

	// JobStep
	Condition        string         `actions:"condition,omitempty"`
	ContinueOnError  *TemplateToken `actions:"continueOnError,omitempty"`
	TimeoutInMinutes *TemplateToken `actions:"timeoutInMinutes,omitempty"`

	// ActionStep
	Reference        StepReference  `actions:"reference,omitempty"`
	ContextName      string         `actions:"contextName,omitempty"`
	DisplayNameToken *TemplateToken `actions:"displayNameToken,omitempty"`
	Env              *TemplateToken `actions:"environment,omitempty"`
	Inputs           *TemplateToken `actions:"inputs,omitempty"`
}

// https://github.com/actions/runner/blob/main/src/Sdk/DTPipelines/Pipelines/ActionStepDefinitionReference.cs
type StepReference struct {
	Type SourceType `actions:"type,omitempty"`

	//ContainerRegistryReference
	Image string `actions:"image,omitempty"`

	// RepositoryPathReference
	Name           string `actions:"name,omitempty"`
	Ref            string `actions:"ref,omitempty"`
	Path           string `actions:"path,omitempty"`
	RepositoryType string `actions:"repositoryType,omitempty"`
}

type SourceType string

const (
	SourceTypeRepository        SourceType = "repository"
	SourceTypeContainerRegistry SourceType = "containerRegistry"
	SourceTypeScript            SourceType = "script"
)
