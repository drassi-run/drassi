package workflows

import (
	"drassi.run/core/pkg/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOn(t *testing.T) {
	t.Run("single-event", func(t *testing.T) {
		expected := &On{
			Push: &OnPush{},
		}

		input := "push"
		o := new(On)
		err := model.Decode(input, o)
		assert.NoError(t, err)
		assert.EqualValues(t, expected, o)
	})

	t.Run("multiple-event", func(t *testing.T) {
		expected := &On{
			Push: &OnPush{},
			Fork: &OnFork{},
		}

		input := []any{"push", "fork"}
		o := new(On)
		err := model.Decode(input, o)
		assert.NoError(t, err)
		assert.EqualValues(t, expected, o)
	})

	t.Run("activity-type", func(t *testing.T) {
		expected := &On{
			Label: &OnLabel{
				Types: []EventLabelActivity{EventLabelActivityCreated},
			},
			Push: &OnPush{
				Branches: []string{"main"},
			},
			// TODO PageBuild is not null
		}

		input := map[string]any{
			"label": map[string]any{
				"types": []any{"created"},
			},
			"push": map[string]any{
				"branches": []any{"main"},
			},
			"page_build": nil,
		}
		o := new(On)
		err := model.Decode(input, o)
		assert.NoError(t, err)
		assert.EqualValues(t, expected, o)
	})
}

type klass[E any] struct {
	Event *E `actions:"event,omitempty"`
}

func createMockData(types any) map[string]any {
	return map[string]any{
		"event": map[string]any{
			"types": types,
		},
	}
}

func testScenario(t *testing.T, check func(tt *testing.T, data map[string]any, value any), value any) {
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

		if value != "error" {
			assert.NoError(tt, err)
		}

		switch value {
		case "error":
			assert.Error(tt, err)
		case "default":
			assert.EqualValues(tt, obj.Event.defaultActivities(), obj.Event.Types)
		case "empty":
			assert.EqualValues(tt, []EventBranchProtectionRuleActivity{}, obj.Event.Types)
		default:
			assert.EqualValues(tt, value, obj.Event.Types)
		}
	}

	testScenario(t, check, []EventBranchProtectionRuleActivity{
		EventBranchProtectionRuleActivityCreate,
	})
}

func TestOnCheckRun(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnCheckRun]{}
		err := model.Decode(data, &obj)

		if value != "error" {
			assert.NoError(tt, err)
		}

		switch value {
		case "error":
			assert.Error(tt, err)
		case "default":
			assert.EqualValues(tt, obj.Event.defaultActivities(), obj.Event.Types)
		case "empty":
			assert.EqualValues(tt, []EventCheckRunActivity{}, obj.Event.Types)
		default:
			assert.EqualValues(tt, value, obj.Event.Types)
		}
	}

	testScenario(t, check, []EventCheckRunActivity{
		EventCheckRunActivityCreated,
	})
}

func TestOnCheckSuite(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnCheckSuite]{}
		err := model.Decode(data, &obj)

		if value != "error" {
			assert.NoError(tt, err)
		}

		switch value {
		case "error":
			assert.Error(tt, err)
		case "default":
			assert.EqualValues(tt, obj.Event.defaultActivities(), obj.Event.Types)
		case "empty":
			assert.EqualValues(tt, []EventCheckSuiteActivity{}, obj.Event.Types)
		default:
			assert.EqualValues(tt, value, obj.Event.Types)
		}
	}

	testScenario(t, check, []EventCheckSuiteActivity{
		EventCheckSuiteActivityCompleted,
	})
}

func TestOnDiscussion(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnDiscussion]{}
		err := model.Decode(data, &obj)

		if value != "error" {
			assert.NoError(tt, err)
		}

		switch value {
		case "error":
			assert.Error(tt, err)
		case "default":
			assert.EqualValues(tt, obj.Event.defaultActivities(), obj.Event.Types)
		case "empty":
			assert.EqualValues(tt, []EventDiscussionActivity{}, obj.Event.Types)
		default:
			assert.EqualValues(tt, value, obj.Event.Types)
		}
	}

	testScenario(t, check, []EventDiscussionActivity{
		EventDiscussionActivityCreated,
	})
}

func TestOnDiscussionComment(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnDiscussionComment]{}
		err := model.Decode(data, &obj)

		if value != "error" {
			assert.NoError(tt, err)
		}

		switch value {
		case "error":
			assert.Error(tt, err)
		case "default":
			assert.EqualValues(tt, obj.Event.defaultActivities(), obj.Event.Types)
		case "empty":
			assert.EqualValues(tt, []EventDiscussionCommentActivity{}, obj.Event.Types)
		default:
			assert.EqualValues(tt, value, obj.Event.Types)
		}
	}

	testScenario(t, check, []EventDiscussionCommentActivity{
		EventDiscussionCommentActivityCreated,
	})
}

func TestOnIssueComment(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnIssueComment]{}
		err := model.Decode(data, &obj)

		if value != "error" {
			assert.NoError(tt, err)
		}

		switch value {
		case "error":
			assert.Error(tt, err)
		case "default":
			assert.EqualValues(tt, obj.Event.defaultActivities(), obj.Event.Types)
		case "empty":
			assert.EqualValues(tt, []EventIssueCommentActivity{}, obj.Event.Types)
		default:
			assert.EqualValues(tt, value, obj.Event.Types)
		}
	}

	testScenario(t, check, []EventIssueCommentActivity{
		EventIssueCommentActivityCreated,
	})
}

func TestOnIssues(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnIssues]{}
		err := model.Decode(data, &obj)

		if value != "error" {
			assert.NoError(tt, err)
		}

		switch value {
		case "error":
			assert.Error(tt, err)
		case "default":
			assert.EqualValues(tt, obj.Event.defaultActivities(), obj.Event.Types)
		case "empty":
			assert.EqualValues(tt, []EventIssuesActivity{}, obj.Event.Types)
		default:
			assert.EqualValues(tt, value, obj.Event.Types)
		}
	}

	testScenario(t, check, []EventIssuesActivity{
		EventIssuesActivityOpened,
	})
}

func TestOnLabel(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnLabel]{}
		err := model.Decode(data, &obj)

		if value != "error" {
			assert.NoError(tt, err)
		}

		switch value {
		case "error":
			assert.Error(tt, err)
		case "default":
			assert.EqualValues(tt, obj.Event.defaultActivities(), obj.Event.Types)
		case "empty":
			assert.EqualValues(tt, []EventLabelActivity{}, obj.Event.Types)
		default:
			assert.EqualValues(tt, value, obj.Event.Types)
		}
	}

	testScenario(t, check, []EventLabelActivity{
		EventLabelActivityCreated,
	})
}

func TestOnMergeGroup(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnMergeGroup]{}
		err := model.Decode(data, &obj)

		if value != "error" {
			assert.NoError(tt, err)
		}

		switch value {
		case "error":
			assert.Error(tt, err)
		case "default":
			assert.EqualValues(tt, obj.Event.defaultActivities(), obj.Event.Types)
		case "empty":
			assert.EqualValues(tt, []EventMergeGroupActivity{}, obj.Event.Types)
		default:
			assert.EqualValues(tt, value, obj.Event.Types)
		}
	}

	testScenario(t, check, []EventMergeGroupActivity{
		EventMergeGroupActivityChecksRequested,
	})
}

func TestOnMilestone(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnMilestone]{}
		err := model.Decode(data, &obj)

		if value != "error" {
			assert.NoError(tt, err)
		}

		switch value {
		case "error":
			assert.Error(tt, err)
		case "default":
			assert.EqualValues(tt, obj.Event.defaultActivities(), obj.Event.Types)
		case "empty":
			assert.EqualValues(tt, []EventMilestoneActivity{}, obj.Event.Types)
		default:
			assert.EqualValues(tt, value, obj.Event.Types)
		}
	}

	testScenario(t, check, []EventMilestoneActivity{
		EventMilestoneActivityCreated,
	})
}

func TestOnProject(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnProject]{}
		err := model.Decode(data, &obj)

		if value != "error" {
			assert.NoError(tt, err)
		}

		switch value {
		case "error":
			assert.Error(tt, err)
		case "default":
			assert.EqualValues(tt, obj.Event.defaultActivities(), obj.Event.Types)
		case "empty":
			assert.EqualValues(tt, []EventProjectActivity{}, obj.Event.Types)
		default:
			assert.EqualValues(tt, value, obj.Event.Types)
		}
	}

	testScenario(t, check, []EventProjectActivity{
		EventProjectActivityReopened,
	})
}

func TestOnProjectCard(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnProjectCard]{}
		err := model.Decode(data, &obj)

		if value != "error" {
			assert.NoError(tt, err)
		}

		switch value {
		case "error":
			assert.Error(tt, err)
		case "default":
			assert.EqualValues(tt, obj.Event.defaultActivities(), obj.Event.Types)
		case "empty":
			assert.EqualValues(tt, []EventProjectCardActivity{}, obj.Event.Types)
		default:
			assert.EqualValues(tt, value, obj.Event.Types)
		}
	}

	testScenario(t, check, []EventProjectCardActivity{
		EventProjectCardActivityCreated,
	})
}

func TestOnProjectColumn(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnProjectColumn]{}
		err := model.Decode(data, &obj)

		if value != "error" {
			assert.NoError(tt, err)
		}

		switch value {
		case "error":
			assert.Error(tt, err)
		case "default":
			assert.EqualValues(tt, obj.Event.defaultActivities(), obj.Event.Types)
		case "empty":
			assert.EqualValues(tt, []EventProjectColumnActivity{}, obj.Event.Types)
		default:
			assert.EqualValues(tt, value, obj.Event.Types)
		}
	}

	testScenario(t, check, []EventProjectColumnActivity{
		EventProjectColumnActivityMoved,
	})
}

func TestOnPullRequest(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnPullRequest]{}
		err := model.Decode(data, &obj)

		if value != "error" {
			assert.NoError(tt, err)
		}

		switch value {
		case "error":
			assert.Error(tt, err)
		case "default":
			assert.EqualValues(tt, obj.Event.defaultActivities(), obj.Event.Types)
		case "empty":
			assert.EqualValues(tt, []EventPullRequestActivity{}, obj.Event.Types)
		default:
			assert.EqualValues(tt, value, obj.Event.Types)
		}
	}

	testScenario(t, check, []EventPullRequestActivity{
		EventPullRequestActivityOpened,
	})
}

func TestOnPullRequestReview(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnPullRequestReview]{}
		err := model.Decode(data, &obj)

		if value != "error" {
			assert.NoError(tt, err)
		}

		switch value {
		case "error":
			assert.Error(tt, err)
		case "default":
			assert.EqualValues(tt, obj.Event.defaultActivities(), obj.Event.Types)
		case "empty":
			assert.EqualValues(tt, []EventPullRequestReviewActivity{}, obj.Event.Types)
		default:
			assert.EqualValues(tt, value, obj.Event.Types)
		}
	}

	testScenario(t, check, []EventPullRequestReviewActivity{
		EventPullRequestReviewActivitySubmitted,
	})
}

func TestOnPullRequestReviewComment(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnPullRequestReviewComment]{}
		err := model.Decode(data, &obj)

		if value != "error" {
			assert.NoError(tt, err)
		}

		switch value {
		case "error":
			assert.Error(tt, err)
		case "default":
			assert.EqualValues(tt, obj.Event.defaultActivities(), obj.Event.Types)
		case "empty":
			assert.EqualValues(tt, []EventPullRequestReviewCommentActivity{}, obj.Event.Types)
		default:
			assert.EqualValues(tt, value, obj.Event.Types)
		}
	}

	testScenario(t, check, []EventPullRequestReviewCommentActivity{
		EventPullRequestReviewCommentActivityCreated,
	})
}

func TestOnPullRequestTarget(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnPullRequestTarget]{}
		err := model.Decode(data, &obj)

		if value != "error" {
			assert.NoError(tt, err)
		}

		switch value {
		case "error":
			assert.Error(tt, err)
		case "default":
			assert.EqualValues(tt, obj.Event.defaultActivities(), obj.Event.Types)
		case "empty":
			assert.EqualValues(tt, []EventPullRequestTargetActivity{}, obj.Event.Types)
		default:
			assert.EqualValues(tt, value, obj.Event.Types)
		}
	}

	testScenario(t, check, []EventPullRequestTargetActivity{
		EventPullRequestTargetActivityOpened,
	})
}

func TestOnRegistryPackage(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnRegistryPackage]{}
		err := model.Decode(data, &obj)

		if value != "error" {
			assert.NoError(tt, err)
		}

		switch value {
		case "error":
			assert.Error(tt, err)
		case "default":
			assert.EqualValues(tt, obj.Event.defaultActivities(), obj.Event.Types)
		case "empty":
			assert.EqualValues(tt, []EventRegistryPackageActivity{}, obj.Event.Types)
		default:
			assert.EqualValues(tt, value, obj.Event.Types)
		}
	}

	testScenario(t, check, []EventRegistryPackageActivity{
		EventRegistryPackageActivityPublished,
	})
}

func TestOnRelease(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnRelease]{}
		err := model.Decode(data, &obj)

		if value != "error" {
			assert.NoError(tt, err)
		}

		switch value {
		case "error":
			assert.Error(tt, err)
		case "default":
			assert.EqualValues(tt, obj.Event.defaultActivities(), obj.Event.Types)
		case "empty":
			assert.EqualValues(tt, []EventReleaseActivity{}, obj.Event.Types)
		default:
			assert.EqualValues(tt, value, obj.Event.Types)
		}
	}

	testScenario(t, check, []EventReleaseActivity{
		EventReleaseActivityPublished,
	})
}

func TestOnWatch(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnWatch]{}
		err := model.Decode(data, &obj)

		if value != "error" {
			assert.NoError(tt, err)
		}

		switch value {
		case "error":
			assert.Error(tt, err)
		case "default":
			assert.EqualValues(tt, obj.Event.defaultActivities(), obj.Event.Types)
		case "empty":
			assert.EqualValues(tt, []EventWatchActivity{}, obj.Event.Types)
		default:
			assert.EqualValues(tt, value, obj.Event.Types)
		}
	}

	testScenario(t, check, []EventWatchActivity{
		EventWatchActivityStarted,
	})
}

func TestOnWorkflowRun(t *testing.T) {
	var check = func(tt *testing.T, data map[string]any, value any) {
		obj := klass[OnWorkflowRun]{}
		err := model.Decode(data, &obj)

		if value != "error" {
			assert.NoError(tt, err)
		}

		switch value {
		case "error":
			assert.Error(tt, err)
		case "default":
			assert.EqualValues(tt, obj.Event.defaultActivities(), obj.Event.Types)
		case "empty":
			assert.EqualValues(tt, []EventWorkflowRunActivity{}, obj.Event.Types)
		default:
			assert.EqualValues(tt, value, obj.Event.Types)
		}
	}

	testScenario(t, check, []EventWorkflowRunActivity{
		EventWorkflowRunActivityCompleted,
	})
}
