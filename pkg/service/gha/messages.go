package gha

import (
	"time"
)

const (
	MessageTypeAgentRefresh            = "AgentRefresh"
	MessageTypeRunnerRefresh           = "RunnerRefresh"
	MessageTypeRunnerShutdown          = "RunnerShutdown"
	MessageTypeJobCancelMessage        = "JobCancellation"
	MessageTypeRunnerJobRequest        = "RunnerJobRequest"
	MessageTypePipelineAgentJobRequest = "PipelineAgentJobRequest"
	MessageTypeForceTokenRefresh       = "ForceTokenRefresh"
)

type Duration struct {
	time.Duration
}

type Time struct {
	time.Time
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/AgentRefreshMessage.cs
type AgentRefreshMessage struct {
	AgentId       int32    `json:"agentId,omitempty"`
	Timeout       Duration `json:"timeout,omitempty"`
	TargetVersion string   `json:"targetVersion,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/RunnerRefreshMessage.cs
type RunnerRefreshMessage struct {
	TargetVersion  string `json:"target_version,omitempty"`
	DownloadUrl    string `json:"download_url,omitempty"`
	SHA256Checksum string `json:"sha256_checksum,omitempty"`
	OS             string `json:"os,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/RunnerShutdownMessage.cs
type RunnerShutdownMessage struct {
	Reason string `json:"reason,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/JobCancelMessage.cs
type JobCancelMessage struct {
	JobId   string   `json:"jobId,omitempty"`
	Timeout Duration `json:"timeout,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Listener/RunnerJobRequestRef.cs
type RunnerJobRequestMessage struct {
	Id              string `json:"id,omitempty"`
	RunnerRequestId string `json:"runner_request_id,omitempty"`
	RunServiceUrl   string `json:"run_service_url,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/AgentJobRequestMessage.cs
type PipelineAgentJobRequestMessage struct {
	MessageType          string                   `json:"messageType,omitempty"`
	RequestId            int64                    `json:"requestId,omitempty"`
	Plan                 PlanReference            `json:"plan,omitempty"`
	Timeline             TimelineReference        `json:"timeline,omitempty"`
	JobId                string                   `json:"jobId,omitempty"`
	JobName              string                   `json:"jobName,omitempty"`
	JobDisplayName       string                   `json:"jobDisplayName,omitempty"`
	JobContainer         *TemplateToken           `json:"jobContainer,omitempty"`
	JobServiceContainers *TemplateToken           `json:"jobServiceContainers,omitempty"`
	JobOutputs           *TemplateToken           `json:"jobOutputs,omitempty"`
	LockedUntil          Time                     `json:"lockedUntil,omitempty"`
	Resources            *JobResources            `json:"resources,omitempty"`
	ContextData          ContextData              `json:"contextData,omitempty"`
	Workspace            *WorkspaceOptions        `json:"workspace,omitempty"`
	MaskHints            []MaskHint               `json:"mask,omitempty"`
	EnvironmentVariables []TemplateToken          `json:"environmentVariables,omitempty"`
	Defaults             []TemplateToken          `json:"defaults,omitempty"`
	Environment          *EnvironmentReference    `json:"actionsEnvironment,omitempty"`
	Snapshot             *TemplateToken           `json:"snapshot,omitempty"`
	Variables            map[string]VariableValue `json:"variables,omitempty"`
	Steps                []JobStep                `json:"steps,omitempty"`
	FileTable            []string                 `json:"fileTable,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/TaskOrchestrationPlanReference.cs
type PlanReference struct {
	ScopeIdentifier  string `json:"scopeIdentifier,omitempty"` // UUID
	PlanType         string `json:"planType,omitempty"`
	Version          int32  `json:"version,omitempty"`
	PlanId           string `json:"planId,omitempty"` // UUID
	PlanGroup        string `json:"planGroup,omitempty"`
	ArtifactUri      string `json:"artifactUri,omitempty"`      // URI
	ArtifactLocation string `json:"artifactLocation,omitempty"` // URI
	ContainerId      int64  `json:"containerId,omitempty"`
	Definition       Owner  `json:"definition,omitempty"`
	Owner            Owner  `json:"owner,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/TaskOrchestrationOwner.cs
type Owner struct {
	Id    int32          `json:"id,omitempty"`
	Name  string         `json:"name,omitempty"`
	Links map[string]any `json:"_links,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/TimelineReference.cs
type TimelineReference struct {
	Id       string `json:"id,omitempty"` // UUID
	ChangeId int32  `json:"changeId,omitempty"`
	Location string `json:"location,omitempty"` // URI
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTObjectTemplating/ObjectTemplating/Tokens/TemplateToken.cs
type TemplateToken struct {
	File   int32 `json:"file,omitempty"`
	Line   int32 `json:"line,omitempty"`
	Column int32 `json:"column,omitempty"`
	Type   int32 `json:"type,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/JobResources.cs
type JobResources struct {
	Endpoints    []ServiceEndpoint    `json:"endpoints,omitempty"`
	Containers   []ContainerResource  `json:"containers,omitempty"`
	Repositories []RepositoryResource `json:"repositories,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/ServiceEndpointLegacy/ServiceEndpoint.cs
type ServiceEndpoint struct {
	Id              string                `json:"id,omitempty"` // UUID
	Name            string                `json:"name,omitempty"`
	Type            string                `json:"type,omitempty"`
	Owner           string                `json:"owner,omitempty"`
	Url             string                `json:"url,omitempty"` // URI
	Description     string                `json:"description,omitempty"`
	Authorization   EndpointAuthorization `json:"authorization,omitempty"`
	GroupScopeId    string                `json:"groupScopeId,omitempty"` // UUID
	Data            map[string]string     `json:"data,omitempty"`
	IsShared        bool                  `json:"isShared,omitempty"`
	IsReady         bool                  `json:"isReady,omitempty"`
	OperationStatus any                   `json:"operationStatus,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/ServiceEndpointLegacy/EndpointAuthorization.cs
type EndpointAuthorization struct {
	Scheme     string            `json:"scheme,omitempty"`
	Parameters map[string]string `json:"parameters,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/Resource.cs
type Resource struct {
	Alias      string                   `json:"alias,omitempty"`
	Endpoint   ServiceEndpointReference `json:"endpoint,omitempty"`
	Properties ResourceProperties       `json:"properties,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/ServiceEndpointReference.cs
type ServiceEndpointReference struct {
	Id   string `json:"id,omitempty"`   // UUID
	Name string `json:"name,omitempty"` // ExpressionValue<String>
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/ResourceProperties.cs
type ResourceProperties struct {
	Count int32          `json:"count,omitempty"`
	Items map[string]any `json:"items,omitempty"` // IDictionary<String, JToken>
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/ContainerResource.cs
type ContainerResource struct {
	Resource `json:",inline"`

	Image   string            `json:"image,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	Ports   []string          `json:"ports,omitempty"`
	Volumes []string          `json:"volumes,omitempty"`
	Options string            `json:"options,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/RepositoryResource.cs
type RepositoryResource struct {
	Resource `json:",inline"`

	Id      string `json:"id,omitempty"`
	Type    string `json:"type,omitempty"`
	Url     string `json:"url,omitempty"` // URI
	Version string `json:"version,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/WorkspaceOptions.cs
type WorkspaceOptions struct {
	Clean string `json:"clean,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/MaskHint.cs
type MaskHint struct {
	Type  MaskType `json:"type,omitempty"`
	Value string   `json:"value,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/MaskType.cs
type MaskType string

const (
	MaskTypeRegex    MaskType = "regex"
	MaskTypeVariable MaskType = "variable"
)

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/ActionsEnvironmentReference.cs
type EnvironmentReference struct {
	Name string        `json:"name,omitempty"`
	Url  TemplateToken `json:"url,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/VariableValue.cs
type VariableValue struct {
	Value    string `json:"value,omitempty"`
	IsSecret bool   `json:"isSecret,omitempty"`
}

type ContextData map[string]any

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/JobStep.cs
type JobStep struct {
}
