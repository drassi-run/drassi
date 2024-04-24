package workflows

import (
	"github.com/dungdm93/drasi/pkg/model"
	"github.com/mitchellh/mapstructure"
	"gotest.tools/v3/assert"
	"testing"
)

type klass[E any] struct {
	Event *E `mapstructure:"event,omitempty"`
}

func createMockData(types any) map[string]any {
	return map[string]any{
		"event": map[string]any{
			"types": types,
		},
	}
}

func testSenerio(t *testing.T, check func(tt *testing.T, data map[string]any, value any), value any) {
	t.Run("absent", func(tt *testing.T) {
		data := map[string]any{
			"event": map[string]any{},
		}
		check(tt, data, "default")
	})

	t.Run("nil", func(tt *testing.T) {
		data := createMockData(nil)
		check(tt, data, "default")
	})

	t.Run("emptyList", func(tt *testing.T) {
		data := createMockData([]string{})
		check(tt, data, "empty")
	})

	t.Run("specifyValue", func(tt *testing.T) {
		data := createMockData(value)
		check(tt, data, value)
	})

	t.Run("err", func(tt *testing.T) {
		data := createMockData("a string")
		check(tt, data, "error")
	})
}

func TestOnBranchProtectionRule(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnBranchProtectionRule]{}
		err := model.Decode(data, &obj)

		if value == "error" {
			assert.ErrorType(tt, err, &mapstructure.Error{})
			return
		}
		assert.NilError(tt, err)

		switch value {
		case "default":
			assert.DeepEqual(tt, obj.Event.Types, obj.Event.defaultActivities())
		case "empty":
			assert.DeepEqual(tt, obj.Event.Types, []EventBranchProtectionRuleActivity{})
		default:
			assert.DeepEqual(tt, obj.Event.Types, value)
		}
	}

	testSenerio(t, check, []EventBranchProtectionRuleActivity{
		EventBranchProtectionRuleActivityCreate,
	})
}

func TestOnCheckRun(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnCheckRun]{}
		err := model.Decode(data, &obj)

		if value == "error" {
			assert.ErrorType(tt, err, &mapstructure.Error{})
			return
		}
		assert.NilError(tt, err)

		switch value {
		case "default":
			assert.DeepEqual(tt, obj.Event.Types, obj.Event.defaultActivities())
		case "empty":
			assert.DeepEqual(tt, obj.Event.Types, []EventCheckRunActivity{})
		default:
			assert.DeepEqual(tt, obj.Event.Types, value)
		}
	}

	testSenerio(t, check, []EventCheckRunActivity{
		EventCheckRunActivityCreated,
	})
}

func TestOnCheckSuite(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnCheckSuite]{}
		err := model.Decode(data, &obj)

		if value == "error" {
			assert.ErrorType(tt, err, &mapstructure.Error{})
			return
		}
		assert.NilError(tt, err)

		switch value {
		case "default":
			assert.DeepEqual(tt, obj.Event.Types, obj.Event.defaultActivities())
		case "empty":
			assert.DeepEqual(tt, obj.Event.Types, []EventCheckSuiteActivity{})
		default:
			assert.DeepEqual(tt, obj.Event.Types, value)
		}
	}

	testSenerio(t, check, []EventCheckSuiteActivity{
		EventCheckSuiteActivityCompleted,
	})
}

func TestOnDiscussion(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnDiscussion]{}
		err := model.Decode(data, &obj)

		if value == "error" {
			assert.ErrorType(tt, err, &mapstructure.Error{})
			return
		}
		assert.NilError(tt, err)

		switch value {
		case "default":
			assert.DeepEqual(tt, obj.Event.Types, obj.Event.defaultActivities())
		case "empty":
			assert.DeepEqual(tt, obj.Event.Types, []EventDiscussionActivity{})
		default:
			assert.DeepEqual(tt, obj.Event.Types, value)
		}
	}

	testSenerio(t, check, []EventDiscussionActivity{
		EventDiscussionActivityCreated,
	})
}

func TestOnDiscussionComment(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnDiscussionComment]{}
		err := model.Decode(data, &obj)

		if value == "error" {
			assert.ErrorType(tt, err, &mapstructure.Error{})
			return
		}
		assert.NilError(tt, err)

		switch value {
		case "default":
			assert.DeepEqual(tt, obj.Event.Types, obj.Event.defaultActivities())
		case "empty":
			assert.DeepEqual(tt, obj.Event.Types, []EventDiscussionCommentActivity{})
		default:
			assert.DeepEqual(tt, obj.Event.Types, value)
		}
	}

	testSenerio(t, check, []EventDiscussionCommentActivity{
		EventDiscussionCommentActivityCreated,
	})
}

func TestOnIssueComment(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnIssueComment]{}
		err := model.Decode(data, &obj)

		if value == "error" {
			assert.ErrorType(tt, err, &mapstructure.Error{})
			return
		}
		assert.NilError(tt, err)

		switch value {
		case "default":
			assert.DeepEqual(tt, obj.Event.Types, obj.Event.defaultActivities())
		case "empty":
			assert.DeepEqual(tt, obj.Event.Types, []EventIssueCommentActivity{})
		default:
			assert.DeepEqual(tt, obj.Event.Types, value)
		}
	}

	testSenerio(t, check, []EventIssueCommentActivity{
		EventIssueCommentActivityCreated,
	})
}

func TestOnIssues(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnIssues]{}
		err := model.Decode(data, &obj)

		if value == "error" {
			assert.ErrorType(tt, err, &mapstructure.Error{})
			return
		}
		assert.NilError(tt, err)

		switch value {
		case "default":
			assert.DeepEqual(tt, obj.Event.Types, obj.Event.defaultActivities())
		case "empty":
			assert.DeepEqual(tt, obj.Event.Types, []EventIssuesActivity{})
		default:
			assert.DeepEqual(tt, obj.Event.Types, value)
		}
	}

	testSenerio(t, check, []EventIssuesActivity{
		EventIssuesActivityOpened,
	})
}

func TestOnLabel(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnLabel]{}
		err := model.Decode(data, &obj)

		if value == "error" {
			assert.ErrorType(tt, err, &mapstructure.Error{})
			return
		}
		assert.NilError(tt, err)

		switch value {
		case "default":
			assert.DeepEqual(tt, obj.Event.Types, obj.Event.defaultActivities())
		case "empty":
			assert.DeepEqual(tt, obj.Event.Types, []EventLabelActivity{})
		default:
			assert.DeepEqual(tt, obj.Event.Types, value)
		}
	}

	testSenerio(t, check, []EventLabelActivity{
		EventLabelActivityCreated,
	})
}

func TestOnMergeGroup(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnMergeGroup]{}
		err := model.Decode(data, &obj)

		if value == "error" {
			assert.ErrorType(tt, err, &mapstructure.Error{})
			return
		}
		assert.NilError(tt, err)

		switch value {
		case "default":
			assert.DeepEqual(tt, obj.Event.Types, obj.Event.defaultActivities())
		case "empty":
			assert.DeepEqual(tt, obj.Event.Types, []EventMergeGroupActivity{})
		default:
			assert.DeepEqual(tt, obj.Event.Types, value)
		}
	}

	testSenerio(t, check, []EventMergeGroupActivity{
		EventMergeGroupActivityChecksRequested,
	})
}

func TestOnMilestone(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnMilestone]{}
		err := model.Decode(data, &obj)

		if value == "error" {
			assert.ErrorType(tt, err, &mapstructure.Error{})
			return
		}
		assert.NilError(tt, err)

		switch value {
		case "default":
			assert.DeepEqual(tt, obj.Event.Types, obj.Event.defaultActivities())
		case "empty":
			assert.DeepEqual(tt, obj.Event.Types, []EventMilestoneActivity{})
		default:
			assert.DeepEqual(tt, obj.Event.Types, value)
		}
	}

	testSenerio(t, check, []EventMilestoneActivity{
		EventMilestoneActivityCreated,
	})
}

func TestOnProject(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnProject]{}
		err := model.Decode(data, &obj)

		if value == "error" {
			assert.ErrorType(tt, err, &mapstructure.Error{})
			return
		}
		assert.NilError(tt, err)

		switch value {
		case "default":
			assert.DeepEqual(tt, obj.Event.Types, obj.Event.defaultActivities())
		case "empty":
			assert.DeepEqual(tt, obj.Event.Types, []EventProjectActivity{})
		default:
			assert.DeepEqual(tt, obj.Event.Types, value)
		}
	}

	testSenerio(t, check, []EventProjectActivity{
		EventProjectActivityReopened,
	})
}

func TestOnProjectCard(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnProjectCard]{}
		err := model.Decode(data, &obj)

		if value == "error" {
			assert.ErrorType(tt, err, &mapstructure.Error{})
			return
		}
		assert.NilError(tt, err)

		switch value {
		case "default":
			assert.DeepEqual(tt, obj.Event.Types, obj.Event.defaultActivities())
		case "empty":
			assert.DeepEqual(tt, obj.Event.Types, []EventProjectCardActivity{})
		default:
			assert.DeepEqual(tt, obj.Event.Types, value)
		}
	}

	testSenerio(t, check, []EventProjectCardActivity{
		EventProjectCardActivityCreated,
	})
}

func TestOnProjectColumn(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnProjectColumn]{}
		err := model.Decode(data, &obj)

		if value == "error" {
			assert.ErrorType(tt, err, &mapstructure.Error{})
			return
		}
		assert.NilError(tt, err)

		switch value {
		case "default":
			assert.DeepEqual(tt, obj.Event.Types, obj.Event.defaultActivities())
		case "empty":
			assert.DeepEqual(tt, obj.Event.Types, []EventProjectColumnActivity{})
		default:
			assert.DeepEqual(tt, obj.Event.Types, value)
		}
	}

	testSenerio(t, check, []EventProjectColumnActivity{
		EventProjectColumnActivityMoved,
	})
}

func TestOnPullRequest(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnPullRequest]{}
		err := model.Decode(data, &obj)

		if value == "error" {
			assert.ErrorType(tt, err, &mapstructure.Error{})
			return
		}
		assert.NilError(tt, err)

		switch value {
		case "default":
			assert.DeepEqual(tt, obj.Event.Types, obj.Event.defaultActivities())
		case "empty":
			assert.DeepEqual(tt, obj.Event.Types, []EventPullRequestActivity{})
		default:
			assert.DeepEqual(tt, obj.Event.Types, value)
		}
	}

	testSenerio(t, check, []EventPullRequestActivity{
		EventPullRequestActivityOpened,
	})
}

func TestOnPullRequestReview(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnPullRequestReview]{}
		err := model.Decode(data, &obj)

		if value == "error" {
			assert.ErrorType(tt, err, &mapstructure.Error{})
			return
		}
		assert.NilError(tt, err)

		switch value {
		case "default":
			assert.DeepEqual(tt, obj.Event.Types, obj.Event.defaultActivities())
		case "empty":
			assert.DeepEqual(tt, obj.Event.Types, []EventPullRequestReviewActivity{})
		default:
			assert.DeepEqual(tt, obj.Event.Types, value)
		}
	}

	testSenerio(t, check, []EventPullRequestReviewActivity{
		EventPullRequestReviewActivitySubmitted,
	})
}

func TestOnPullRequestReviewComment(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnPullRequestReviewComment]{}
		err := model.Decode(data, &obj)

		if value == "error" {
			assert.ErrorType(tt, err, &mapstructure.Error{})
			return
		}
		assert.NilError(tt, err)

		switch value {
		case "default":
			assert.DeepEqual(tt, obj.Event.Types, obj.Event.defaultActivities())
		case "empty":
			assert.DeepEqual(tt, obj.Event.Types, []EventPullRequestReviewCommentActivity{})
		default:
			assert.DeepEqual(tt, obj.Event.Types, value)
		}
	}

	testSenerio(t, check, []EventPullRequestReviewCommentActivity{
		EventPullRequestReviewCommentActivityCreated,
	})
}

func TestOnPullRequestTarget(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnPullRequestTarget]{}
		err := model.Decode(data, &obj)

		if value == "error" {
			assert.ErrorType(tt, err, &mapstructure.Error{})
			return
		}
		assert.NilError(tt, err)

		switch value {
		case "default":
			assert.DeepEqual(tt, obj.Event.Types, obj.Event.defaultActivities())
		case "empty":
			assert.DeepEqual(tt, obj.Event.Types, []EventPullRequestTargetActivity{})
		default:
			assert.DeepEqual(tt, obj.Event.Types, value)
		}
	}

	testSenerio(t, check, []EventPullRequestTargetActivity{
		EventPullRequestTargetActivityOpened,
	})
}

func TestOnRegistryPackage(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnRegistryPackage]{}
		err := model.Decode(data, &obj)

		if value == "error" {
			assert.ErrorType(tt, err, &mapstructure.Error{})
			return
		}
		assert.NilError(tt, err)

		switch value {
		case "default":
			assert.DeepEqual(tt, obj.Event.Types, obj.Event.defaultActivities())
		case "empty":
			assert.DeepEqual(tt, obj.Event.Types, []EventRegistryPackageActivity{})
		default:
			assert.DeepEqual(tt, obj.Event.Types, value)
		}
	}

	testSenerio(t, check, []EventRegistryPackageActivity{
		EventRegistryPackageActivityPublished,
	})
}

func TestOnRelease(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnRelease]{}
		err := model.Decode(data, &obj)

		if value == "error" {
			assert.ErrorType(tt, err, &mapstructure.Error{})
			return
		}
		assert.NilError(tt, err)

		switch value {
		case "default":
			assert.DeepEqual(tt, obj.Event.Types, obj.Event.defaultActivities())
		case "empty":
			assert.DeepEqual(tt, obj.Event.Types, []EventReleaseActivity{})
		default:
			assert.DeepEqual(tt, obj.Event.Types, value)
		}
	}

	testSenerio(t, check, []EventReleaseActivity{
		EventReleaseActivityPublished,
	})
}

func TestOnWatch(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnWatch]{}
		err := model.Decode(data, &obj)

		if value == "error" {
			assert.ErrorType(tt, err, &mapstructure.Error{})
			return
		}
		assert.NilError(tt, err)

		switch value {
		case "default":
			assert.DeepEqual(tt, obj.Event.Types, obj.Event.defaultActivities())
		case "empty":
			assert.DeepEqual(tt, obj.Event.Types, []EventWatchActivity{})
		default:
			assert.DeepEqual(tt, obj.Event.Types, value)
		}
	}

	testSenerio(t, check, []EventWatchActivity{
		EventWatchActivityStarted,
	})
}

func TestOnWorkflowRun(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnWorkflowRun]{}
		err := model.Decode(data, &obj)

		if value == "error" {
			assert.ErrorType(tt, err, &mapstructure.Error{})
			return
		}
		assert.NilError(tt, err)

		switch value {
		case "default":
			assert.DeepEqual(tt, obj.Event.Types, obj.Event.defaultActivities())
		case "empty":
			assert.DeepEqual(tt, obj.Event.Types, []EventWorkflowRunActivity{})
		default:
			assert.DeepEqual(tt, obj.Event.Types, value)
		}
	}

	testSenerio(t, check, []EventWorkflowRunActivity{
		EventWorkflowRunActivityCompleted,
	})
}
