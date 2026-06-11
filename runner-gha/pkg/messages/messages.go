/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package messages

import (
	"bytes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"time"

	"drassi.run/gha-runner/pkg/types"
	"golang.org/x/oauth2"
)

const (
	TypeBrokerMigration         = "BrokerMigration"
	TypeAgentRefresh            = "AgentRefresh"
	TypeRunnerRefresh           = "RunnerRefresh"
	TypeRunnerShutdown          = "RunnerShutdown"
	TypeRunnerRefreshConfig     = "RunnerRefreshConfig"
	TypeJobCancellation         = "JobCancellation"
	TypeRunnerJobRequest        = "RunnerJobRequest"
	TypePipelineAgentJobRequest = "PipelineAgentJobRequest"
	TypeForceTokenRefresh       = "ForceTokenRefresh"
)

// Message provides a contract for receiving messages from the task orchestrator.
// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/TaskAgentMessage.cs
type Message struct {
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

func (m *Message) IsEmpty() bool {
	if m == nil {
		return true
	}
	return m.Id == 0 && m.Type == ""
}

func (m *Message) DecryptBody(key cipher.Block) ([]byte, error) {
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

// BrokerMigration is Message that tells the runner to redirect itself to BrokerListener for messages.
// (Note that we use a special Message instead of a simple 302. This is because
// the runner will need to apply the runner's token to the request, and it is
// a security best practice to *not* blindly add sensitive data to redirects
// 302s.)
// https://github.com/actions/runner/blob/v2.323.0/src/Sdk/DTWebApi/WebApi/BrokerMigrationMessage.cs
// https://github.com/actions/runner/pull/3103
type BrokerMigration struct {
	// The base url for the broker listener
	BaseUrl string `json:"brokerBaseUrl,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/AgentRefreshMessage.cs
type AgentRefresh struct {
	AgentId       int32         `json:"agentId,omitempty"`
	Timeout       time.Duration `json:"timeout,omitempty"`
	TargetVersion string        `json:"targetVersion,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/RunnerRefreshMessage.cs
type RunnerRefresh struct {
	TargetVersion  string `json:"target_version,omitempty"`
	DownloadUrl    string `json:"download_url,omitempty"`
	SHA256Checksum string `json:"sha256_checksum,omitempty"`
	OS             string `json:"os,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/RunnerShutdownMessage.cs
type RunnerShutdown struct {
	Reason string `json:"reason,omitempty"`
}

// https://github.com/actions/runner/blob/v2.335.1/src/Sdk/DTWebApi/WebApi/RunnerRefreshConfigMessage.cs
type RunnerRefreshConfig struct {
	RunnerQualifiedId string `json:"runner_qualified_id,omitempty"`
	ConfigType        string `json:"config_type,omitempty"`
	ServiceType       string `json:"service_type,omitempty"`
	ConfigRefreshUrl  string `json:"config_refresh_url,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/JobCancelMessage.cs
type JobCancel struct {
	JobId   string        `json:"jobId,omitempty"`
	Timeout time.Duration `json:"timeout,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Listener/RunnerJobRequestRef.cs
type RunnerJobRequest struct {
	Id              string `json:"id,omitempty"`
	RunnerRequestId string `json:"runner_request_id,omitempty"`
	RunServiceUrl   string `json:"run_service_url,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/AgentJobRequestMessage.cs
type PipelineAgentJobRequest struct {
	MessageType          string                `json:"messageType,omitempty"`
	RequestId            int64                 `json:"requestId,omitempty"`
	Plan                 PlanReference         `json:"plan,omitempty"`
	Timeline             TimelineReference     `json:"timeline,omitempty"`
	JobId                string                `json:"jobId,omitempty"`
	JobName              string                `json:"jobName,omitempty"`
	JobDisplayName       string                `json:"jobDisplayName,omitempty"`
	JobContainer         *TemplateToken        `json:"jobContainer,omitempty"`
	JobServiceContainers *TemplateToken        `json:"jobServiceContainers,omitempty"`
	JobOutputs           *TemplateToken        `json:"jobOutputs,omitempty"`
	LockedUntil          time.Time             `json:"lockedUntil,omitempty"`
	Resources            *JobResources         `json:"resources,omitempty"`
	ContextData          map[string]Value      `json:"contextData,omitempty"`
	Workspace            *WorkspaceOptions     `json:"workspace,omitempty"`
	MaskHints            []MaskHint            `json:"mask,omitempty"`
	Env                  []TemplateToken       `json:"environmentVariables,omitempty"`
	Defaults             []TemplateToken       `json:"defaults,omitempty"`
	Environment          *EnvironmentReference `json:"actionsEnvironment,omitempty"`
	Snapshot             *TemplateToken        `json:"snapshot,omitempty"`
	Variables            map[string]Variable   `json:"variables,omitempty"`
	Steps                []JobStep             `json:"steps,omitempty"`
	FileTable            []string              `json:"fileTable,omitempty"`
	BillingOwnerId       string                `json:"billing_owner_id,omitempty"`
}

func (m *PipelineAgentJobRequest) ServiceEndpoint(name string) *ServiceEndpoint {
	res := m.Resources
	if res == nil {
		return nil
	}

	for i := range res.Endpoints {
		ep := &res.Endpoints[i]
		if ep.Name == name {
			return ep
		}
	}
	return nil
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
	Type   TokenType `json:"type,omitempty"`
	File   int32     `json:"file,omitempty"`
	Line   int32     `json:"line,omitempty"`
	Column int32     `json:"column,omitempty"`

	String    string                                   `json:"lit,omitempty"`       // StringToken (type=0)
	Number    float64                                  `json:"num,omitempty"`       // NumberToken (type=6)
	Boolean   bool                                     `json:"bool,omitempty"`      // BooleanToken (type=5)
	Directive string                                   `json:"directive,omitempty"` // InsertExpressionToken (type=4)
	Expr      string                                   `json:"expr,omitempty"`      // BasicExpressionToken (type=3)
	Seq       []*TemplateToken                         `json:"seq,omitempty"`       // SequenceToken (type=1)
	Map       []KVPair[*TemplateToken, *TemplateToken] `json:"map,omitempty"`       // MappingToken (type=2)
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

// Github Actions recently change from key/value to Key/Value
type KVPair[K, V any] struct {
	Key   K `json:"key,omitempty,case:ignore"`
	Value V `json:"value,omitempty,case:ignore"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/JobResources.cs
type JobResources struct {
	Endpoints    []ServiceEndpoint    `json:"endpoints,omitempty"`
	Containers   []ContainerResource  `json:"containers,omitempty"`
	Repositories []RepositoryResource `json:"repositories,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTWebApi/WebApi/ServiceEndpointLegacy/ServiceEndpoint.cs
type ServiceEndpoint struct {
	Id              string                 `json:"id,omitempty"` // UUID
	Name            string                 `json:"name,omitempty"`
	Type            string                 `json:"type,omitempty"`
	Owner           string                 `json:"owner,omitempty"`
	Url             string                 `json:"url,omitempty"` // URI
	Description     string                 `json:"description,omitempty"`
	Authorization   *EndpointAuthorization `json:"authorization,omitempty"`
	GroupScopeId    string                 `json:"groupScopeId,omitempty"` // UUID
	Data            map[string]string      `json:"data,omitempty"`
	IsShared        bool                   `json:"isShared,omitempty"`
	IsReady         bool                   `json:"isReady,omitempty"`
	OperationStatus any                    `json:"operationStatus,omitempty"`
}

func (ep *ServiceEndpoint) TokenSource() (oauth2.TokenSource, error) {
	auth := ep.Authorization
	if auth == nil {
		return nil, nil
	}

	if auth.Scheme != "OAuth" {
		return nil, fmt.Errorf("unsupported authorization scheme: %s", auth.Scheme)
	}
	if at, ok := auth.Parameters["AccessToken"]; ok {
		token := &oauth2.Token{
			AccessToken: at,
			TokenType:   "Bearer",
		}
		return oauth2.StaticTokenSource(token), nil
	}

	return nil, nil
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
type Variable struct {
	Value    string `json:"value,omitempty"`
	IsSecret bool   `json:"isSecret,omitempty"`
}

type Value any

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/JobStep.cs
type JobStep struct {
	// Step
	Id          string `json:"id,omitempty"`
	Type        string `json:"type,omitempty"`
	Name        string `json:"name,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	Enabled     bool   `json:"enabled,omitempty"`

	// JobStep
	Condition        string         `json:"condition,omitempty"`
	ContinueOnError  *TemplateToken `json:"continueOnError,omitempty"`
	TimeoutInMinutes *TemplateToken `json:"timeoutInMinutes,omitempty"`

	// ActionStep
	Reference        *StepReference `json:"reference,omitempty"`
	ContextName      string         `json:"contextName,omitempty"`
	DisplayNameToken *TemplateToken `json:"displayNameToken,omitempty"`
	Env              *TemplateToken `json:"environment,omitempty"`
	Inputs           *TemplateToken `json:"inputs,omitempty"`
}

// https://github.com/actions/runner/blob/main/src/Sdk/DTPipelines/Pipelines/ActionStepDefinitionReference.cs
type StepReference struct {
	Type Source `json:"type,omitempty"`

	//ContainerRegistryReference
	Image string `json:"image,omitempty"`

	// RepositoryPathReference
	Name           string `json:"name,omitempty"`
	Ref            string `json:"ref,omitempty"`
	Path           string `json:"path,omitempty"`
	RepositoryType string `json:"repositoryType,omitempty"`
}

type Source string

const (
	SourceRepository        Source = "repository"
	SourceContainerRegistry Source = "containerRegistry"
	SourceScript            Source = "script"
)
