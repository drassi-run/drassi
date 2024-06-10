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
	AgentId       int32    `mapstructure:"agentId,omitempty"`
	Timeout       Duration `mapstructure:"timeout,omitempty"`
	TargetVersion string   `mapstructure:"targetVersion,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/RunnerRefreshMessage.cs
type RunnerRefresh struct {
	TargetVersion  string `mapstructure:"target_version,omitempty"`
	DownloadUrl    string `mapstructure:"download_url,omitempty"`
	SHA256Checksum string `mapstructure:"sha256_checksum,omitempty"`
	OS             string `mapstructure:"os,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/RunnerShutdownMessage.cs
type RunnerShutdown struct {
	Reason string `mapstructure:"reason,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/JobCancelMessage.cs
type JobCancel struct {
	JobId   string   `mapstructure:"jobId,omitempty"`
	Timeout Duration `mapstructure:"timeout,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Listener/RunnerJobRequestRef.cs
type RunnerJobRequest struct {
	Id              string `mapstructure:"id,omitempty"`
	RunnerRequestId string `mapstructure:"runner_request_id,omitempty"`
	RunServiceUrl   string `mapstructure:"run_service_url,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/AgentJobRequestMessage.cs
type PipelineAgentJobRequest struct {
	MessageType          string                   `mapstructure:"messageType,omitempty"`
	RequestId            int64                    `mapstructure:"requestId,omitempty"`
	Plan                 PlanReference            `mapstructure:"plan,omitempty"`
	Timeline             TimelineReference        `mapstructure:"timeline,omitempty"`
	JobId                string                   `mapstructure:"jobId,omitempty"`
	JobName              string                   `mapstructure:"jobName,omitempty"`
	JobDisplayName       string                   `mapstructure:"jobDisplayName,omitempty"`
	JobContainer         *TemplateToken           `mapstructure:"jobContainer,omitempty"`
	JobServiceContainers *TemplateToken           `mapstructure:"jobServiceContainers,omitempty"`
	JobOutputs           *TemplateToken           `mapstructure:"jobOutputs,omitempty"`
	LockedUntil          Time                     `mapstructure:"lockedUntil,omitempty"`
	Resources            *JobResources            `mapstructure:"resources,omitempty"`
	ContextData          ContextData              `mapstructure:"contextData,omitempty"`
	Workspace            *WorkspaceOptions        `mapstructure:"workspace,omitempty"`
	MaskHints            []MaskHint               `mapstructure:"mask,omitempty"`
	Env                  []TemplateToken          `mapstructure:"environmentVariables,omitempty"`
	Defaults             []TemplateToken          `mapstructure:"defaults,omitempty"`
	Environment          *EnvironmentReference    `mapstructure:"actionsEnvironment,omitempty"`
	Snapshot             *TemplateToken           `mapstructure:"snapshot,omitempty"`
	Variables            map[string]VariableValue `mapstructure:"variables,omitempty"`
	Steps                []JobStep                `mapstructure:"steps,omitempty"`
	FileTable            []string                 `mapstructure:"fileTable,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/TaskOrchestrationPlanReference.cs
type PlanReference struct {
	ScopeIdentifier  string `mapstructure:"scopeIdentifier,omitempty"` // UUID
	PlanType         string `mapstructure:"planType,omitempty"`
	Version          int32  `mapstructure:"version,omitempty"`
	PlanId           string `mapstructure:"planId,omitempty"` // UUID
	PlanGroup        string `mapstructure:"planGroup,omitempty"`
	ArtifactUri      string `mapstructure:"artifactUri,omitempty"`      // URI
	ArtifactLocation string `mapstructure:"artifactLocation,omitempty"` // URI
	ContainerId      int64  `mapstructure:"containerId,omitempty"`
	Definition       Owner  `mapstructure:"definition,omitempty"`
	Owner            Owner  `mapstructure:"owner,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/TaskOrchestrationOwner.cs
type Owner struct {
	Id    int32          `mapstructure:"id,omitempty"`
	Name  string         `mapstructure:"name,omitempty"`
	Links map[string]any `mapstructure:"_links,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/TimelineReference.cs
type TimelineReference struct {
	Id       string `mapstructure:"id,omitempty"` // UUID
	ChangeId int32  `mapstructure:"changeId,omitempty"`
	Location string `mapstructure:"location,omitempty"` // URI
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTObjectTemplating/ObjectTemplating/Tokens/TemplateToken.cs
type TemplateToken struct {
	Type   TokenType `mapstructure:"type,omitempty"`
	File   int32     `mapstructure:"file,omitempty"`
	Line   int32     `mapstructure:"line,omitempty"`
	Column int32     `mapstructure:"column,omitempty"`

	String    string                                   `mapstructure:"lit,omitempty"`       // StringToken (type=0)
	Number    float64                                  `mapstructure:"num,omitempty"`       // NumberToken (type=6)
	Boolean   bool                                     `mapstructure:"bool,omitempty"`      // BooleanToken (type=5)
	Directive string                                   `mapstructure:"directive,omitempty"` // InsertExpressionToken (type=4)
	Expr      string                                   `mapstructure:"expr,omitempty"`      // BasicExpressionToken (type=3)
	Seq       []*TemplateToken                         `mapstructure:"seq,omitempty"`       // SequenceToken (type=1)
	Map       []KVPair[*TemplateToken, *TemplateToken] `mapstructure:"map,omitempty"`       // MappingToken (type=2)
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
	Key   K `mapstructure:"key,omitempty"`
	Value V `mapstructure:"value,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/JobResources.cs
type JobResources struct {
	Endpoints    []ServiceEndpoint    `mapstructure:"endpoints,omitempty"`
	Containers   []ContainerResource  `mapstructure:"containers,omitempty"`
	Repositories []RepositoryResource `mapstructure:"repositories,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/ServiceEndpointLegacy/ServiceEndpoint.cs
type ServiceEndpoint struct {
	Id              string                `mapstructure:"id,omitempty"` // UUID
	Name            string                `mapstructure:"name,omitempty"`
	Type            string                `mapstructure:"type,omitempty"`
	Owner           string                `mapstructure:"owner,omitempty"`
	Url             string                `mapstructure:"url,omitempty"` // URI
	Description     string                `mapstructure:"description,omitempty"`
	Authorization   EndpointAuthorization `mapstructure:"authorization,omitempty"`
	GroupScopeId    string                `mapstructure:"groupScopeId,omitempty"` // UUID
	Data            map[string]string     `mapstructure:"data,omitempty"`
	IsShared        bool                  `mapstructure:"isShared,omitempty"`
	IsReady         bool                  `mapstructure:"isReady,omitempty"`
	OperationStatus any                   `mapstructure:"operationStatus,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/ServiceEndpointLegacy/EndpointAuthorization.cs
type EndpointAuthorization struct {
	Scheme     string            `mapstructure:"scheme,omitempty"`
	Parameters map[string]string `mapstructure:"parameters,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/Resource.cs
type Resource struct {
	Alias      string                   `mapstructure:"alias,omitempty"`
	Endpoint   ServiceEndpointReference `mapstructure:"endpoint,omitempty"`
	Properties ResourceProperties       `mapstructure:"properties,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/ServiceEndpointReference.cs
type ServiceEndpointReference struct {
	Id   string `mapstructure:"id,omitempty"`   // UUID
	Name string `mapstructure:"name,omitempty"` // ExpressionValue<String>
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/ResourceProperties.cs
type ResourceProperties struct {
	Count int32          `mapstructure:"count,omitempty"`
	Items map[string]any `mapstructure:"items,omitempty"` // IDictionary<String, JToken>
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/ContainerResource.cs
type ContainerResource struct {
	Resource `mapstructure:",inline"`

	Image   string            `mapstructure:"image,omitempty"`
	Env     map[string]string `mapstructure:"env,omitempty"`
	Ports   []string          `mapstructure:"ports,omitempty"`
	Volumes []string          `mapstructure:"volumes,omitempty"`
	Options string            `mapstructure:"options,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/RepositoryResource.cs
type RepositoryResource struct {
	Resource `mapstructure:",inline"`

	Id      string `mapstructure:"id,omitempty"`
	Type    string `mapstructure:"type,omitempty"`
	Url     string `mapstructure:"url,omitempty"` // URI
	Version string `mapstructure:"version,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/WorkspaceOptions.cs
type WorkspaceOptions struct {
	Clean string `mapstructure:"clean,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/MaskHint.cs
type MaskHint struct {
	Type  MaskType `mapstructure:"type,omitempty"`
	Value string   `mapstructure:"value,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/MaskType.cs
type MaskType string

const (
	MaskTypeRegex    MaskType = "regex"
	MaskTypeVariable MaskType = "variable"
)

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/ActionsEnvironmentReference.cs
type EnvironmentReference struct {
	Name string        `mapstructure:"name,omitempty"`
	Url  TemplateToken `mapstructure:"url,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/VariableValue.cs
type VariableValue struct {
	Value    string `mapstructure:"value,omitempty"`
	IsSecret bool   `mapstructure:"isSecret,omitempty"`
}

type ContextData map[string]any

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/JobStep.cs
type JobStep struct {
	// Step
	Id          string `mapstructure:"id,omitempty"`
	Type        string `mapstructure:"type,omitempty"`
	Name        string `mapstructure:"name,omitempty"`
	DisplayName string `mapstructure:"displayName,omitempty"`
	Enabled     bool   `mapstructure:"enabled,omitempty"`

	// JobStep
	Condition        string         `mapstructure:"condition,omitempty"`
	ContinueOnError  *TemplateToken `mapstructure:"continueOnError,omitempty"`
	TimeoutInMinutes *TemplateToken `mapstructure:"timeoutInMinutes,omitempty"`

	// ActionStep
	Reference        StepReference  `mapstructure:"reference,omitempty"`
	ContextName      string         `mapstructure:"contextName,omitempty"`
	DisplayNameToken *TemplateToken `mapstructure:"displayNameToken,omitempty"`
	Env              *TemplateToken `mapstructure:"environment,omitempty"`
	Inputs           *TemplateToken `mapstructure:"inputs,omitempty"`
}

// https://github.com/actions/runner/blob/main/src/Sdk/DTPipelines/Pipelines/ActionStepDefinitionReference.cs
type StepReference struct {
	Type SourceType `mapstructure:"type,omitempty"`

	//ContainerRegistryReference
	Image string `mapstructure:"image,omitempty"`

	// RepositoryPathReference
	Name           string `mapstructure:"name,omitempty"`
	Ref            string `mapstructure:"ref,omitempty"`
	Path           string `mapstructure:"path,omitempty"`
	RepositoryType string `mapstructure:"repositoryType,omitempty"`
}

type SourceType string

const (
	SourceTypeRepository        SourceType = "repository"
	SourceTypeContainerRegistry SourceType = "containerRegistry"
	SourceTypeScript            SourceType = "script"
)
