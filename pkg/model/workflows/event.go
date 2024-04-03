package workflows

type empty = struct{}
type onEventWithTypes[T ~string] struct {
	Types []T `json:"types,omitempty" yaml:"types,omitempty" mapstructure:"types,omitempty"`
}
type onEventWithRef struct {
	Branches       []string `json:"branches,omitempty" yaml:"branches,omitempty" mapstructure:"branches,omitempty"`
	Tags           []string `json:"tags,omitempty" yaml:"tags,omitempty" mapstructure:"tags,omitempty"`
	Paths          []string `json:"paths,omitempty" yaml:"paths,omitempty" mapstructure:"paths,omitempty"`
	BranchesIgnore []string `json:"branches-ignore,omitempty" yaml:"branches-ignore,omitempty" mapstructure:"branches-ignore,omitempty"`
	TagsIgnore     []string `json:"tags-ignore,omitempty" yaml:"tags-ignore,omitempty" mapstructure:"tags-ignore,omitempty"`
	PathsIgnore    []string `json:"paths-ignore,omitempty" yaml:"paths-ignore,omitempty" mapstructure:"paths-ignore,omitempty"`
}

// The name of the GitHub event that triggers the workflow.
// You can provide a single event string, array of events, array of event types, or an event configuration map
// that schedules a workflow or restricts the execution of a workflow to specific files, tags, or branch changes.
// For a list of available events, see https://help.github.com/en/github/automating-your-workflow-with-github-actions/events-that-trigger-workflows.
// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#on
type On struct {
	// Runs your workflow anytime the branch_protection_rule event occurs. More than one activity type triggers this event.
	// https://docs.github.com/en/actions/learn-github-actions/events-that-trigger-workflows#branch_protection_rule
	BranchProtectionRule *OnBranchProtectionRule `json:"branch_protection_rule,omitempty" yaml:"branch_protection_rule,omitempty" mapstructure:"branch_protection_rule,omitempty"`

	// Runs your workflow anytime the check_run event occurs. More than one activity type triggers this event.
	// For information about the REST API, see https://developer.github.com/v3/checks/runs.
	// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#check_run
	CheckRun *OnCheckRun `json:"check_run,omitempty" yaml:"check_run,omitempty" mapstructure:"check_run,omitempty"`

	// Runs your workflow anytime the check_suite event occurs. More than one activity type triggers this event.
	// For information about the REST API, see https://developer.github.com/v3/checks/suites/.
	// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#check_suite
	CheckSuite *OnCheckSuite `json:"check_suite,omitempty" yaml:"check_suite,omitempty" mapstructure:"check_suite,omitempty"`

	// Runs your workflow anytime someone creates a branch or tag, which triggers the create event.
	// For information about the REST API, see https://developer.github.com/v3/git/refs/#create-a-reference.
	// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#create
	Create *OnCreate `json:"create,omitempty" yaml:"create,omitempty" mapstructure:"create,omitempty"`

	// Runs your workflow anytime someone deletes a branch or tag, which triggers the delete event.
	// For information about the REST API, see https://developer.github.com/v3/git/refs/#delete-a-reference.
	// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#delete
	Delete *OnDelete `json:"delete,omitempty" yaml:"delete,omitempty" mapstructure:"delete,omitempty"`

	// Runs your workflow anytime someone creates a deployment, which triggers the deployment event.
	// Deployments created with a commit SHA may not have a Git ref.
	// For information about the REST API, see https://developer.github.com/v3/repos/deployments/.
	// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#deployment
	Deployment *OnDeployment `json:"deployment,omitempty" yaml:"deployment,omitempty" mapstructure:"deployment,omitempty"`

	// Runs your workflow anytime a third party provides a deployment status, which triggers the deployment_status event.
	// Deployments created with a commit SHA may not have a Git ref.
	// For information about the REST API, see https://developer.github.com/v3/repos/deployments/#create-a-deployment-status.
	// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#deployment_status
	DeploymentStatus *OnDeploymentStatus `json:"deployment_status,omitempty" yaml:"deployment_status,omitempty" mapstructure:"deployment_status,omitempty"`

	// Runs your workflow anytime the discussion event occurs. More than one activity type triggers this event.
	// For information about the GraphQL API, see https://docs.github.com/en/graphql/guides/using-the-graphql-api-for-discussions
	// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#discussion
	Discussion *OnDiscussion `json:"discussion,omitempty" yaml:"discussion,omitempty" mapstructure:"discussion,omitempty"`

	// Runs your workflow anytime the discussion_comment event occurs. More than one activity type triggers this event.
	// For information about the GraphQL API, see https://docs.github.com/en/graphql/guides/using-the-graphql-api-for-discussions
	// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#discussion_comment
	DiscussionComment *OnDiscussionComment `json:"discussion_comment,omitempty" yaml:"discussion_comment,omitempty" mapstructure:"discussion_comment,omitempty"`

	// Runs your workflow anytime when someone forks a repository, which triggers the fork event.
	// For information about the REST API, see https://developer.github.com/v3/repos/forks/#create-a-fork.
	// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#fork
	Fork *OnFork `json:"fork,omitempty" yaml:"fork,omitempty" mapstructure:"fork,omitempty"`

	// Runs your workflow when someone creates or updates a Wiki page, which triggers the gollum event.
	// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#gollum
	Gollum *OnGollum `json:"gollum,omitempty" yaml:"gollum,omitempty" mapstructure:"gollum,omitempty"`

	// Runs your workflow anytime the issue_comment event occurs. More than one activity type triggers this event.
	// For information about the REST API, see https://developer.github.com/v3/issues/comments/.
	// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#issue_comment
	IssueComment *OnIssueComment `json:"issue_comment,omitempty" yaml:"issue_comment,omitempty" mapstructure:"issue_comment,omitempty"`

	// Runs your workflow anytime the issues event occurs. More than one activity type triggers this event.
	// For information about the REST API, see https://developer.github.com/v3/issues.
	// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#issues
	Issues *OnIssues `json:"issues,omitempty" yaml:"issues,omitempty" mapstructure:"issues,omitempty"`

	// Runs your workflow anytime the label event occurs. More than one activity type triggers this event.
	// For information about the REST API, see https://developer.github.com/v3/issues/labels/.
	// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#label
	Label *OnLabel `json:"label,omitempty" yaml:"label,omitempty" mapstructure:"label,omitempty"`

	// Runs your workflow when a pull request is added to a merge queue, which adds the pull request to a merge group.
	// For information about the merge queue, see https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/incorporating-changes-from-a-pull-request/merging-a-pull-request-with-a-merge-queue.
	// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#merge_group
	MergeGroup *OnMergeGroup `json:"merge_group,omitempty" yaml:"merge_group,omitempty" mapstructure:"merge_group,omitempty"`

	// Runs your workflow anytime the milestone event occurs. More than one activity type triggers this event.
	// For information about the REST API, see https://developer.github.com/v3/issues/milestones/.
	// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#milestone
	Milestone *OnMilestone `json:"milestone,omitempty" yaml:"milestone,omitempty" mapstructure:"milestone,omitempty"`

	// Runs your workflow anytime someone pushes to a GitHub Pages-enabled branch, which triggers the page_build event.
	// For information about the REST API, see https://developer.github.com/v3/repos/pages/.
	// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#page_build
	PageBuild *OnPageBuild `json:"page_build,omitempty" yaml:"page_build,omitempty" mapstructure:"page_build,omitempty"`

	// Runs your workflow anytime the project event occurs. More than one activity type triggers this event.
	// For information about the REST API, see https://developer.github.com/v3/projects/.
	// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#project
	Project *OnProject `json:"project,omitempty" yaml:"project,omitempty" mapstructure:"project,omitempty"`

	// Runs your workflow anytime the project_card event occurs. More than one activity type triggers this event.
	// For information about the REST API, see https://developer.github.com/v3/projects/cards.
	// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#project_card
	ProjectCard *OnProjectCard `json:"project_card,omitempty" yaml:"project_card,omitempty" mapstructure:"project_card,omitempty"`

	// Runs your workflow anytime the project_column event occurs. More than one activity type triggers this event.
	// For information about the REST API, see https://developer.github.com/v3/projects/columns.
	// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#project_column
	ProjectColumn *OnProjectColumn `json:"project_column,omitempty" yaml:"project_column,omitempty" mapstructure:"project_column,omitempty"`

	// Runs your workflow anytime someone makes a private repository public, which triggers the public event.
	// For information about the REST API, see https://developer.github.com/v3/repos/#edit.
	// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#public
	Public *OnPublic `json:"public,omitempty" yaml:"public,omitempty" mapstructure:"public,omitempty"`

	// Runs your workflow anytime the pull_request event occurs. More than one activity type triggers this event.
	// For information about the REST API, see https://developer.github.com/v3/pulls.
	// Note: Workflows do not run on private base repositories when you open a pull request from a forked repository.
	// When you create a pull request from a forked repository to the base repository, GitHub sends
	// the pull_request event to the base repository and no pull request events occur on the forked repository.
	// Workflows don't run on forked repositories by default. You must enable GitHub Actions in the Actions tab
	// of the forked repository. The permissions for the GITHUB_TOKEN in forked repositories is read-only.
	// For more information about the GITHUB_TOKEN, see https://help.github.com/en/articles/virtual-environments-for-github-actions.
	// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#pull_request
	PullRequest *OnPullRequest `json:"pull_request,omitempty" yaml:"pull_request,omitempty" mapstructure:"pull_request,omitempty"`

	// Runs your workflow anytime the pull_request_review event occurs. More than one activity type triggers this event.
	// For information about the REST API, see https://developer.github.com/v3/pulls/reviews.
	// Note: Workflows do not run on private base repositories when you open a pull request from a forked repository.
	// When you create a pull request from a forked repository to the base repository, GitHub sends
	// the pull_request event to the base repository and no pull request events occur on the forked repository.
	// Workflows don't run on forked repositories by default. You must enable GitHub Actions in the Actions tab
	// of the forked repository. The permissions for the GITHUB_TOKEN in forked repositories is read-only.
	// For more information about the GITHUB_TOKEN, see https://help.github.com/en/articles/virtual-environments-for-github-actions.
	// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#pull_request_review
	PullRequestReview *OnPullRequestReview `json:"pull_request_review,omitempty" yaml:"pull_request_review,omitempty" mapstructure:"pull_request_review,omitempty"`

	// Runs your workflow anytime a comment on a pull request's unified diff is modified, which triggers the pull_request_review_comment event.
	// More than one activity type triggers this event. For information about the REST API, see https://developer.github.com/v3/pulls/comments.
	// Note: Workflows do not run on private base repositories when you open a pull request from a forked repository.
	// When you create a pull request from a forked repository to the base repository, GitHub sends the pull_request
	// event to the base repository and no pull request events occur on the forked repository.
	// Workflows don't run on forked repositories by default. You must enable GitHub Actions in the Actions tab
	// of the forked repository. The permissions for the GITHUB_TOKEN in forked repositories is read-only.
	// For more information about the GITHUB_TOKEN, see https://help.github.com/en/articles/virtual-environments-for-github-actions.
	// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#pull_request_review_comment
	PullRequestReviewComment *OnPullRequestReviewComment `json:"pull_request_review_comment,omitempty" yaml:"pull_request_review_comment,omitempty" mapstructure:"pull_request_review_comment,omitempty"`

	// This event is similar to pull_request, except that it runs in the context of the base repository of the pull request, rather than in the merge commit.
	// This means that you can more safely make your secrets available to the workflows triggered by the pull request,
	// because only workflows defined in the commit on the base repository are run.
	// For example, this event allows you to create workflows that label and comment on pull requests, based on the contents of the event payload.
	// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#pull_request_target
	PullRequestTarget *OnPullRequestTarget `json:"pull_request_target,omitempty" yaml:"pull_request_target,omitempty" mapstructure:"pull_request_target,omitempty"`

	// Runs your workflow when someone pushes to a repository branch, which triggers the push event.
	// Note: The webhook payload available to GitHub Actions does not include the added, removed, and
	// modified attributes in the commit object. You can retrieve the full commit object using the REST API.
	// For more information, see https://developer.github.com/v3/repos/commits/#get-a-single-commit.
	// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#push
	Push *OnPush `json:"push,omitempty" yaml:"push,omitempty" mapstructure:"push,omitempty"`

	// Runs your workflow anytime a package is published or updated.
	// For more information, see https://help.github.com/en/github/managing-packages-with-github-packages.
	// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#registry_package
	RegistryPackage *OnRegistryPackage `json:"registry_package,omitempty" yaml:"registry_package,omitempty" mapstructure:"registry_package,omitempty"`

	// Runs your workflow anytime the release event occurs. More than one activity type triggers this event.
	// For information about the REST API, see https://developer.github.com/v3/repos/releases/ in the GitHub Developer documentation.
	// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#release
	Release *OnRelease `json:"release,omitempty" yaml:"release,omitempty" mapstructure:"release,omitempty"`

	// You can use the GitHub API to trigger a webhook event called repository_dispatch when you want to trigger
	// a workflow for activity that happens outside of GitHub.
	// For more information, see https://developer.github.com/v3/repos/#create-a-repository-dispatch-event.
	// To trigger the custom repository_dispatch webhook event, you must send a POST request to a GitHub API endpoint and
	// provide an event_type name to describe the activity type.
	// To trigger a workflow run, you must also configure your workflow to use the repository_dispatch event.
	// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#repository_dispatch
	RepositoryDispatch *OnRepositoryDispatch `json:"repository_dispatch,omitempty" yaml:"repository_dispatch,omitempty" mapstructure:"repository_dispatch,omitempty"`

	// You can schedule a workflow to run at specific UTC times using POSIX cron syntax
	// (https://pubs.opengroup.org/onlinepubs/9699919799/utilities/crontab.html#tag_20_25_07).
	// Scheduled workflows run on the latest commit on the default or base branch.
	// The shortest interval you can run scheduled workflows is once every 5 minutes.
	// Note: GitHub Actions does not support the non-standard syntax @yearly, @monthly, @weekly, @daily, @hourly, and @reboot.
	// You can use crontab guru (https://crontab.guru/). to help generate your cron syntax and confirm what time it will run.
	// To help you get started, there is also a list of crontab guru examples (https://crontab.guru/examples.html).
	// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#schedule
	Schedule *OnSchedule `json:"schedule,omitempty" yaml:"schedule,omitempty" mapstructure:"schedule,omitempty"`

	// Runs your workflow anytime the status of a Git commit changes, which triggers the status event.
	// For information about the REST API, see https://developer.github.com/v3/repos/statuses/.
	// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#status
	Status *OnStatus `json:"status,omitempty" yaml:"status,omitempty" mapstructure:"status,omitempty"`

	// Runs your workflow anytime the watch event occurs. More than one activity type triggers this event.
	// For information about the REST API, see https://developer.github.com/v3/activity/starring/.
	// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#watch
	Watch *OnWatch `json:"watch,omitempty" yaml:"watch,omitempty" mapstructure:"watch,omitempty"`

	// Allows workflows to be reused by other workflows.
	// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#workflow_call
	WorkflowCall *OnWorkflowCall `json:"workflow_call,omitempty" yaml:"workflow_call,omitempty" mapstructure:"workflow_call,omitempty"`

	// You can now create workflows that are manually triggered with the new workflow_dispatch event.
	// You will then see a 'Run workflow' button on the Actions tab, enabling you to easily trigger a run.
	// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#workflow_dispatch
	WorkflowDispatch *OnWorkflowDispatch `json:"workflow_dispatch,omitempty" yaml:"workflow_dispatch,omitempty" mapstructure:"workflow_dispatch,omitempty"`

	// This event occurs when a workflow run is requested or completed, and allows you to execute a workflow
	// based on the finished result of another workflow. For example, if your pull_request workflow generates build artifacts,
	// you can create a new workflow that uses workflow_run to analyze the results and add a comment to the original pull request.
	// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#workflow_run
	WorkflowRun *OnWorkflowRun `json:"workflow_run,omitempty" yaml:"workflow_run,omitempty" mapstructure:"workflow_run,omitempty"`
}

// Event https://help.github.com/en/github/automating-your-workflow-with-github-actions/events-that-trigger-workflows
type Event string

const (
	EventBranchProtectionRule     Event = "branch_protection_rule"
	EventCheckRun                 Event = "check_run"
	EventCheckSuite               Event = "check_suite"
	EventCreate                   Event = "create"
	EventDelete                   Event = "delete"
	EventDeployment               Event = "deployment"
	EventDeploymentStatus         Event = "deployment_status"
	EventDiscussion               Event = "discussion"
	EventDiscussionComment        Event = "discussion_comment"
	EventFork                     Event = "fork"
	EventGollum                   Event = "gollum"
	EventIssueComment             Event = "issue_comment"
	EventIssues                   Event = "issues"
	EventLabel                    Event = "label"
	EventMergeGroup               Event = "merge_group"
	EventMilestone                Event = "milestone"
	EventPageBuild                Event = "page_build"
	EventProject                  Event = "project"
	EventProjectCard              Event = "project_card"
	EventProjectColumn            Event = "project_column"
	EventPublic                   Event = "public"
	EventPullRequest              Event = "pull_request"
	EventPullRequestReview        Event = "pull_request_review"
	EventPullRequestReviewComment Event = "pull_request_review_comment"
	EventPullRequestTarget        Event = "pull_request_target"
	EventPush                     Event = "push"
	EventRegistryPackage          Event = "registry_package"
	EventRelease                  Event = "release"
	EventRepositoryDispatch       Event = "repository_dispatch"
	EventSchedule                 Event = "schedule"
	EventStatus                   Event = "status"
	EventWatch                    Event = "watch"
	EventWorkflowCall             Event = "workflow_call"
	EventWorkflowDispatch         Event = "workflow_dispatch"
	EventWorkflowRun              Event = "workflow_run"
)

type OnBranchProtectionRule onEventWithTypes[EventBranchProtectionRuleActivity]
type EventBranchProtectionRuleActivity string

const (
	EventBranchProtectionRuleActivityCreate  EventBranchProtectionRuleActivity = "created"
	EventBranchProtectionRuleActivityEdited  EventBranchProtectionRuleActivity = "edited"
	EventBranchProtectionRuleActivityDeleted EventBranchProtectionRuleActivity = "deleted"
)

type OnCheckRun onEventWithTypes[EventCheckRunActivity]
type EventCheckRunActivity string

const (
	EventCheckRunActivityCreated         EventCheckRunActivity = "created"
	EventCheckRunActivityRerequested     EventCheckRunActivity = "rerequested"
	EventCheckRunActivityCompleted       EventCheckRunActivity = "completed"
	EventCheckRunActivityRequestedAction EventCheckRunActivity = "requested_action"
)

type OnCheckSuite onEventWithTypes[EventCheckSuiteActivity]
type EventCheckSuiteActivity string

const (
	EventCheckSuiteActivityCompleted   EventCheckSuiteActivity = "completed"
	EventCheckSuiteActivityRequested   EventCheckSuiteActivity = "requested"
	EventCheckSuiteActivityRerequested EventCheckSuiteActivity = "rerequested"
)

type OnCreate = empty
type OnDelete = empty
type OnDeployment = empty
type OnDeploymentStatus = empty

type OnDiscussion onEventWithTypes[EventDiscussionActivity]
type EventDiscussionActivity string

const (
	EventDiscussionActivityCreated         EventDiscussionActivity = "created"
	EventDiscussionActivityEdited          EventDiscussionActivity = "edited"
	EventDiscussionActivityDeleted         EventDiscussionActivity = "deleted"
	EventDiscussionActivityTransferred     EventDiscussionActivity = "transferred"
	EventDiscussionActivityPinned          EventDiscussionActivity = "pinned"
	EventDiscussionActivityUnpinned        EventDiscussionActivity = "unpinned"
	EventDiscussionActivityLabeled         EventDiscussionActivity = "labeled"
	EventDiscussionActivityUnlabeled       EventDiscussionActivity = "unlabeled"
	EventDiscussionActivityLocked          EventDiscussionActivity = "locked"
	EventDiscussionActivityUnlocked        EventDiscussionActivity = "unlocked"
	EventDiscussionActivityCategoryChanged EventDiscussionActivity = "category_changed"
	EventDiscussionActivityAnswered        EventDiscussionActivity = "answered"
	EventDiscussionActivityUnanswered      EventDiscussionActivity = "unanswered"
)

type OnDiscussionComment onEventWithTypes[EventDiscussionCommentActivity]
type EventDiscussionCommentActivity string

const (
	EventDiscussionCommentActivityCreated EventDiscussionCommentActivity = "created"
	EventDiscussionCommentActivityEdited  EventDiscussionCommentActivity = "edited"
	EventDiscussionCommentActivityDeleted EventDiscussionCommentActivity = "deleted"
)

type OnFork = empty
type OnGollum = empty

type OnIssueComment onEventWithTypes[EventIssueCommentActivity]
type EventIssueCommentActivity string

const (
	EventIssueCommentActivityCreated EventIssueCommentActivity = "created"
	EventIssueCommentActivityEdited  EventIssueCommentActivity = "edited"
	EventIssueCommentActivityDeleted EventIssueCommentActivity = "deleted"
)

type OnIssues onEventWithTypes[EventIssuesActivity]
type EventIssuesActivity string

const (
	EventIssuesActivityOpened       EventIssuesActivity = "opened"
	EventIssuesActivityEdited       EventIssuesActivity = "edited"
	EventIssuesActivityDeleted      EventIssuesActivity = "deleted"
	EventIssuesActivityTransferred  EventIssuesActivity = "transferred"
	EventIssuesActivityPinned       EventIssuesActivity = "pinned"
	EventIssuesActivityUnpinned     EventIssuesActivity = "unpinned"
	EventIssuesActivityClosed       EventIssuesActivity = "closed"
	EventIssuesActivityReopened     EventIssuesActivity = "reopened"
	EventIssuesActivityAssigned     EventIssuesActivity = "assigned"
	EventIssuesActivityUnassigned   EventIssuesActivity = "unassigned"
	EventIssuesActivityLabeled      EventIssuesActivity = "labeled"
	EventIssuesActivityUnlabeled    EventIssuesActivity = "unlabeled"
	EventIssuesActivityLocked       EventIssuesActivity = "locked"
	EventIssuesActivityUnlocked     EventIssuesActivity = "unlocked"
	EventIssuesActivityMilestoned   EventIssuesActivity = "milestoned"
	EventIssuesActivityDemilestoned EventIssuesActivity = "demilestoned"
)

type OnLabel onEventWithTypes[EventLabelActivity]
type EventLabelActivity string

const (
	EventLabelActivityCreated EventLabelActivity = "created"
	EventLabelActivityEdited  EventLabelActivity = "edited"
	EventLabelActivityDeleted EventLabelActivity = "deleted"
)

type OnMergeGroup onEventWithTypes[EventMergeGroupActivity]
type EventMergeGroupActivity string

const EventMergeGroupActivityChecksRequested EventMergeGroupActivity = "checks_requested"

type OnMilestone onEventWithTypes[EventMilestoneActivity]
type EventMilestoneActivity string

const (
	EventMilestoneActivityCreated EventMilestoneActivity = "created"
	EventMilestoneActivityClosed  EventMilestoneActivity = "closed"
	EventMilestoneActivityOpened  EventMilestoneActivity = "opened"
	EventMilestoneActivityEdited  EventMilestoneActivity = "edited"
	EventMilestoneActivityDeleted EventMilestoneActivity = "deleted"
)

type OnPageBuild = empty

type OnProject onEventWithTypes[EventProjectActivity]
type EventProjectActivity string

const (
	EventProjectActivityCreated  EventProjectActivity = "created"
	EventProjectActivityClosed   EventProjectActivity = "closed"
	EventProjectActivityReopened EventProjectActivity = "reopened"
	EventProjectActivityEdited   EventProjectActivity = "edited"
	EventProjectActivityDeleted  EventProjectActivity = "deleted"
)

type OnProjectCard onEventWithTypes[EventProjectCardActivity]
type EventProjectCardActivity string

const (
	EventProjectCardActivityCreated   EventProjectCardActivity = "created"
	EventProjectCardActivityMoved     EventProjectCardActivity = "moved"
	EventProjectCardActivityConverted EventProjectCardActivity = "converted" // converted to an issue
	EventProjectCardActivityEdited    EventProjectCardActivity = "edited"
	EventProjectCardActivityDeleted   EventProjectCardActivity = "deleted"
)

type OnProjectColumn onEventWithTypes[EventProjectColumnActivity]
type EventProjectColumnActivity string

const (
	EventProjectColumnActivityCreated EventProjectColumnActivity = "created"
	EventProjectColumnActivityUpdated EventProjectColumnActivity = "updated"
	EventProjectColumnActivityMoved   EventProjectColumnActivity = "moved"
	EventProjectColumnActivityDeleted EventProjectColumnActivity = "deleted"
)

type OnPublic = empty

type OnPullRequest struct {
	onEventWithTypes[EventPullRequestActivity] `yaml:",inline" mapstructure:",squash"`
	onEventWithRef                             `yaml:",inline" mapstructure:",squash"`
}
type EventPullRequestActivity string

const (
	EventPullRequestActivityAssigned             EventPullRequestActivity = "assigned"
	EventPullRequestActivityUnassigned           EventPullRequestActivity = "unassigned"
	EventPullRequestActivityLabeled              EventPullRequestActivity = "labeled"
	EventPullRequestActivityUnlabeled            EventPullRequestActivity = "unlabeled"
	EventPullRequestActivityOpened               EventPullRequestActivity = "opened"
	EventPullRequestActivityEdited               EventPullRequestActivity = "edited"
	EventPullRequestActivityClosed               EventPullRequestActivity = "closed"
	EventPullRequestActivityReopened             EventPullRequestActivity = "reopened"
	EventPullRequestActivitySynchronize          EventPullRequestActivity = "synchronize"
	EventPullRequestActivityConvertedToDraft     EventPullRequestActivity = "converted_to_draft"
	EventPullRequestActivityLocked               EventPullRequestActivity = "locked"
	EventPullRequestActivityUnlocked             EventPullRequestActivity = "unlocked"
	EventPullRequestActivityEnqueued             EventPullRequestActivity = "enqueued"
	EventPullRequestActivityDequeued             EventPullRequestActivity = "dequeued"
	EventPullRequestActivityMilestoned           EventPullRequestActivity = "milestoned"
	EventPullRequestActivityDemilestoned         EventPullRequestActivity = "demilestoned"
	EventPullRequestActivityReadyForReview       EventPullRequestActivity = "ready_for_review"
	EventPullRequestActivityReviewRequested      EventPullRequestActivity = "review_requested"
	EventPullRequestActivityReviewRequestRemoved EventPullRequestActivity = "review_request_removed"
	EventPullRequestActivityAutoMergeEnabled     EventPullRequestActivity = "auto_merge_enabled"
	EventPullRequestActivityAutoMergeDisabled    EventPullRequestActivity = "auto_merge_disabled"
)

type OnPullRequestReview onEventWithTypes[EventPullRequestReviewActivity]
type EventPullRequestReviewActivity string

const (
	EventPullRequestReviewActivitySubmitted EventPullRequestReviewActivity = "submitted"
	EventPullRequestReviewActivityEdited    EventPullRequestReviewActivity = "edited"
	EventPullRequestReviewActivityDismissed EventPullRequestReviewActivity = "dismissed"
)

type OnPullRequestReviewComment onEventWithTypes[EventPullRequestReviewCommentActivity]
type EventPullRequestReviewCommentActivity string

const (
	EventPullRequestReviewCommentActivityCreated EventPullRequestReviewCommentActivity = "created"
	EventPullRequestReviewCommentActivityEdited  EventPullRequestReviewCommentActivity = "edited"
	EventPullRequestReviewCommentActivityDeleted EventPullRequestReviewCommentActivity = "deleted"
)

type OnPullRequestTarget struct {
	onEventWithTypes[EventPullRequestTargetActivity] `yaml:",inline" mapstructure:",squash"`
	onEventWithRef                                   `yaml:",inline" mapstructure:",squash"`
}
type EventPullRequestTargetActivity string

const (
	EventPullRequestTargetActivityAssigned             EventPullRequestTargetActivity = "assigned"
	EventPullRequestTargetActivityUnassigned           EventPullRequestTargetActivity = "unassigned"
	EventPullRequestTargetActivityLabeled              EventPullRequestTargetActivity = "labeled"
	EventPullRequestTargetActivityUnlabeled            EventPullRequestTargetActivity = "unlabeled"
	EventPullRequestTargetActivityOpened               EventPullRequestTargetActivity = "opened"
	EventPullRequestTargetActivityEdited               EventPullRequestTargetActivity = "edited"
	EventPullRequestTargetActivityClosed               EventPullRequestTargetActivity = "closed"
	EventPullRequestTargetActivityReopened             EventPullRequestTargetActivity = "reopened"
	EventPullRequestTargetActivitySynchronize          EventPullRequestTargetActivity = "synchronize"
	EventPullRequestTargetActivityConvertedToDraft     EventPullRequestTargetActivity = "converted_to_draft"
	EventPullRequestTargetActivityReadyForReview       EventPullRequestTargetActivity = "ready_for_review"
	EventPullRequestTargetActivityLocked               EventPullRequestTargetActivity = "locked"
	EventPullRequestTargetActivityUnlocked             EventPullRequestTargetActivity = "unlocked"
	EventPullRequestTargetActivityReviewRequested      EventPullRequestTargetActivity = "review_requested"
	EventPullRequestTargetActivityReviewRequestRemoved EventPullRequestTargetActivity = "review_request_removed"
	EventPullRequestTargetActivityAutoMergeEnabled     EventPullRequestTargetActivity = "auto_merge_enabled"
	EventPullRequestTargetActivityAutoMergeDisabled    EventPullRequestTargetActivity = "auto_merge_disabled"
)

type OnPush onEventWithRef

type OnRegistryPackage onEventWithTypes[EventRegistryPackageActivity]
type EventRegistryPackageActivity string

const (
	EventRegistryPackageActivityPublished EventRegistryPackageActivity = "published"
	EventRegistryPackageActivityUpdated   EventRegistryPackageActivity = "updated"
)

type OnRelease onEventWithTypes[EventReleaseActivity]
type EventReleaseActivity string

const (
	EventReleaseActivityPublished   EventReleaseActivity = "published"
	EventReleaseActivityUnpublished EventReleaseActivity = "unpublished"
	EventReleaseActivityCreated     EventReleaseActivity = "created"
	EventReleaseActivityEdited      EventReleaseActivity = "edited"
	EventReleaseActivityDeleted     EventReleaseActivity = "deleted"
	EventReleaseActivityPrereleased EventReleaseActivity = "prereleased"
	EventReleaseActivityReleased    EventReleaseActivity = "released"
)

type OnRepositoryDispatch onEventWithTypes[EventRepositoryDispatchActivity]
type EventRepositoryDispatchActivity string // any strings

type OnSchedule []struct {
	// https://stackoverflow.com/a/57639657/4044345
	Cron string `json:"cron,omitempty" yaml:"cron,omitempty" mapstructure:"cron,omitempty"`
}
type OnStatus = empty

type OnWatch onEventWithTypes[EventWatchActivity]
type EventWatchActivity string

const EventWatchActivityStarted EventWatchActivity = "started"

type OnWorkflowCall struct {
	// When using the workflow_call keyword, you can optionally specify inputs that are passed to the called workflow from the caller workflow.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#onworkflow_callinputs
	Inputs map[string]Input `json:"inputs,omitempty" yaml:"inputs,omitempty" mapstructure:"inputs,omitempty"`

	// A map of outputs for a called workflow. Called workflow outputs are available to all downstream jobs in the caller workflow.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#onworkflow_calloutputs
	Outputs map[string]Output `json:"outputs,omitempty" yaml:"outputs,omitempty" mapstructure:"outputs,omitempty"`

	// A map of the secrets that can be used in the called workflow.
	// Within the called workflow, you can use the secrets context to refer to a secret.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#onworkflow_callsecrets
	Secrets map[string]Secret `json:"secrets,omitempty" yaml:"secrets,omitempty" mapstructure:"secrets,omitempty"`
}

type OnWorkflowDispatch struct {
	// Input parameters allow you to specify data that the action expects to use during runtime.
	// GitHub stores input parameters as environment variables.
	// Input ids with uppercase letters are converted to lowercase during runtime. We recommended using lowercase input ids.
	// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#inputs
	Inputs map[string]Input `json:"inputs,omitempty" yaml:"inputs,omitempty" mapstructure:"inputs,omitempty"`
}

type OnWorkflowRun struct {
	onEventWithTypes[EventWorkflowRunActivity] `yaml:",inline" mapstructure:",squash"`
	onEventWithRef                             `yaml:",inline" mapstructure:",squash"`

	Workflows []string `json:"workflows,omitempty" yaml:"workflows,omitempty" mapstructure:"workflows,omitempty"`
}
type EventWorkflowRunActivity string

const (
	EventWorkflowRunActivityCompleted  EventWorkflowRunActivity = "completed"
	EventWorkflowRunActivityRequested  EventWorkflowRunActivity = "requested"
	EventWorkflowRunActivityInProgress EventWorkflowRunActivity = "in_progress"
)
