package workflows

func (o *On) DecodeMapstructure(input any) (any, error) {
	var events []string
	switch e := input.(type) {
	case string:
		events = []string{e}
	case []string:
		events = e
	default:
		// process On normal way
		return input, nil
	}

	m := map[string]any{}
	for _, e := range events {
		m[e] = map[string]any{}
	}
	return m, nil
}

func (o *OnBranchProtectionRule) DecodeMapstructure(input any) (any, error) {
	whenEventWithoutTypes(input, func() {
		o.Types = o.defaultActivities()
	})
	return input, nil
}

func (o *OnBranchProtectionRule) defaultActivities() []EventBranchProtectionRuleActivity {
	// all activity types
	return []EventBranchProtectionRuleActivity{
		EventBranchProtectionRuleActivityCreate,
		EventBranchProtectionRuleActivityEdited,
		EventBranchProtectionRuleActivityDeleted,
	}
}

func (o *OnCheckRun) DecodeMapstructure(input any) (any, error) {
	whenEventWithoutTypes(input, func() {
		o.Types = o.defaultActivities()
	})
	return input, nil
}

func (o *OnCheckRun) defaultActivities() []EventCheckRunActivity {
	// all activity types
	return []EventCheckRunActivity{
		EventCheckRunActivityCreated,
		EventCheckRunActivityRerequested,
		EventCheckRunActivityCompleted,
		EventCheckRunActivityRequestedAction,
	}
}

func (o *OnCheckSuite) DecodeMapstructure(input any) (any, error) {
	whenEventWithoutTypes(input, func() {
		o.Types = o.defaultActivities()
	})
	return input, nil
}

func (o *OnCheckSuite) defaultActivities() []EventCheckSuiteActivity {
	// all activity types
	return []EventCheckSuiteActivity{
		EventCheckSuiteActivityCompleted,
		EventCheckSuiteActivityRequested,
		EventCheckSuiteActivityRerequested,
	}
}

func (o *OnDiscussion) DecodeMapstructure(input any) (any, error) {
	whenEventWithoutTypes(input, func() {
		o.Types = o.defaultActivities()
	})
	return input, nil
}
func (o *OnDiscussion) defaultActivities() []EventDiscussionActivity {
	// all activity types
	return []EventDiscussionActivity{
		EventDiscussionActivityCreated,
		EventDiscussionActivityEdited,
		EventDiscussionActivityDeleted,
		EventDiscussionActivityTransferred,
		EventDiscussionActivityPinned,
		EventDiscussionActivityUnpinned,
		EventDiscussionActivityLabeled,
		EventDiscussionActivityUnlabeled,
		EventDiscussionActivityLocked,
		EventDiscussionActivityUnlocked,
		EventDiscussionActivityCategoryChanged,
		EventDiscussionActivityAnswered,
		EventDiscussionActivityUnanswered,
	}
}

func (o *OnDiscussionComment) DecodeMapstructure(input any) (any, error) {
	whenEventWithoutTypes(input, func() {
		o.Types = o.defaultActivities()
	})
	return input, nil
}

func (o *OnDiscussionComment) defaultActivities() []EventDiscussionCommentActivity {
	// all activity types
	return []EventDiscussionCommentActivity{
		EventDiscussionCommentActivityCreated,
		EventDiscussionCommentActivityEdited,
		EventDiscussionCommentActivityDeleted,
	}
}

func (o *OnIssueComment) DecodeMapstructure(input any) (any, error) {
	whenEventWithoutTypes(input, func() {
		o.Types = o.defaultActivities()
	})
	return input, nil
}

func (o *OnIssueComment) defaultActivities() []EventIssueCommentActivity {
	// all activity types
	return []EventIssueCommentActivity{
		EventIssueCommentActivityCreated,
		EventIssueCommentActivityEdited,
		EventIssueCommentActivityDeleted,
	}
}

func (o *OnIssues) DecodeMapstructure(input any) (any, error) {
	whenEventWithoutTypes(input, func() {
		o.Types = o.defaultActivities()
	})
	return input, nil
}

func (o *OnIssues) defaultActivities() []EventIssuesActivity {
	// all activity types
	return []EventIssuesActivity{
		EventIssuesActivityOpened,
		EventIssuesActivityEdited,
		EventIssuesActivityDeleted,
		EventIssuesActivityTransferred,
		EventIssuesActivityPinned,
		EventIssuesActivityUnpinned,
		EventIssuesActivityClosed,
		EventIssuesActivityReopened,
		EventIssuesActivityAssigned,
		EventIssuesActivityUnassigned,
		EventIssuesActivityLabeled,
		EventIssuesActivityUnlabeled,
		EventIssuesActivityLocked,
		EventIssuesActivityUnlocked,
		EventIssuesActivityMilestoned,
		EventIssuesActivityDemilestoned,
	}
}

func (o *OnLabel) DecodeMapstructure(input any) (any, error) {
	whenEventWithoutTypes(input, func() {
		o.Types = o.defaultActivities()
	})
	return input, nil
}

func (o *OnLabel) defaultActivities() []EventLabelActivity {
	// all activity types
	return []EventLabelActivity{
		EventLabelActivityCreated,
		EventLabelActivityEdited,
		EventLabelActivityDeleted,
	}
}

func (o *OnMergeGroup) DecodeMapstructure(input any) (any, error) {
	whenEventWithoutTypes(input, func() {
		o.Types = o.defaultActivities()
	})
	return input, nil
}

func (o *OnMergeGroup) defaultActivities() []EventMergeGroupActivity {
	// all activity types
	return []EventMergeGroupActivity{
		EventMergeGroupActivityChecksRequested,
	}
}

func (o *OnMilestone) DecodeMapstructure(input any) (any, error) {
	whenEventWithoutTypes(input, func() {
		o.Types = o.defaultActivities()
	})
	return input, nil
}

func (o *OnMilestone) defaultActivities() []EventMilestoneActivity {
	// all activity types
	return []EventMilestoneActivity{
		EventMilestoneActivityCreated,
		EventMilestoneActivityClosed,
		EventMilestoneActivityOpened,
		EventMilestoneActivityEdited,
		EventMilestoneActivityDeleted,
	}
}

func (o *OnProject) DecodeMapstructure(input any) (any, error) {
	whenEventWithoutTypes(input, func() {
		o.Types = o.defaultActivities()
	})
	return input, nil
}

func (o *OnProject) defaultActivities() []EventProjectActivity {
	// all activity types
	return []EventProjectActivity{
		EventProjectActivityCreated,
		EventProjectActivityClosed,
		EventProjectActivityReopened,
		EventProjectActivityEdited,
		EventProjectActivityDeleted,
	}
}

func (o *OnProjectCard) DecodeMapstructure(input any) (any, error) {
	whenEventWithoutTypes(input, func() {
		o.Types = o.defaultActivities()
	})
	return input, nil
}

func (o *OnProjectCard) defaultActivities() []EventProjectCardActivity {
	// all activity types
	return []EventProjectCardActivity{
		EventProjectCardActivityCreated,
		EventProjectCardActivityMoved,
		EventProjectCardActivityConverted,
		EventProjectCardActivityEdited,
		EventProjectCardActivityDeleted,
	}
}

func (o *OnProjectColumn) DecodeMapstructure(input any) (any, error) {
	whenEventWithoutTypes(input, func() {
		o.Types = o.defaultActivities()
	})
	return input, nil
}

func (o *OnProjectColumn) defaultActivities() []EventProjectColumnActivity {
	// all activity types
	return []EventProjectColumnActivity{
		EventProjectColumnActivityCreated,
		EventProjectColumnActivityUpdated,
		EventProjectColumnActivityMoved,
		EventProjectColumnActivityDeleted,
	}
}

func (o *OnPullRequest) DecodeMapstructure(input any) (any, error) {
	whenEventWithoutTypes(input, func() {
		o.Types = o.defaultActivities()
	})
	return input, nil
}

func (o *OnPullRequest) defaultActivities() []EventPullRequestActivity {
	return []EventPullRequestActivity{
		EventPullRequestActivityOpened,
		EventPullRequestActivitySynchronize,
		EventPullRequestActivityReopened,
	}
}

func (o *OnPullRequestReview) DecodeMapstructure(input any) (any, error) {
	whenEventWithoutTypes(input, func() {
		o.Types = o.defaultActivities()
	})
	return input, nil
}

func (o *OnPullRequestReview) defaultActivities() []EventPullRequestReviewActivity {
	// all activity types
	return []EventPullRequestReviewActivity{
		EventPullRequestReviewActivitySubmitted,
		EventPullRequestReviewActivityEdited,
		EventPullRequestReviewActivityDismissed,
	}
}

func (o *OnPullRequestReviewComment) DecodeMapstructure(input any) (any, error) {
	whenEventWithoutTypes(input, func() {
		o.Types = o.defaultActivities()
	})
	return input, nil
}

func (o *OnPullRequestReviewComment) defaultActivities() []EventPullRequestReviewCommentActivity {
	// all activity types
	return []EventPullRequestReviewCommentActivity{
		EventPullRequestReviewCommentActivityCreated,
		EventPullRequestReviewCommentActivityEdited,
		EventPullRequestReviewCommentActivityDeleted,
	}
}

func (o *OnPullRequestTarget) DecodeMapstructure(input any) (any, error) {
	whenEventWithoutTypes(input, func() {
		o.Types = o.defaultActivities()
	})
	return input, nil
}

func (o *OnPullRequestTarget) defaultActivities() []EventPullRequestTargetActivity {
	return []EventPullRequestTargetActivity{
		EventPullRequestTargetActivityOpened,
		EventPullRequestTargetActivitySynchronize,
		EventPullRequestTargetActivityReopened,
	}
}

func (o *OnRegistryPackage) DecodeMapstructure(input any) (any, error) {
	whenEventWithoutTypes(input, func() {
		o.Types = o.defaultActivities()
	})
	return input, nil
}

func (o *OnRegistryPackage) defaultActivities() []EventRegistryPackageActivity {
	// all activity types
	return []EventRegistryPackageActivity{
		EventRegistryPackageActivityPublished,
		EventRegistryPackageActivityUpdated,
	}
}

func (o *OnRelease) DecodeMapstructure(input any) (any, error) {
	whenEventWithoutTypes(input, func() {
		o.Types = o.defaultActivities()
	})
	return input, nil
}

func (o *OnRelease) defaultActivities() []EventReleaseActivity {
	// all activity types
	return []EventReleaseActivity{
		EventReleaseActivityPublished,
		EventReleaseActivityUnpublished,
		EventReleaseActivityCreated,
		EventReleaseActivityEdited,
		EventReleaseActivityDeleted,
		EventReleaseActivityPrereleased,
		EventReleaseActivityReleased,
	}
}

func (o *OnWatch) DecodeMapstructure(input any) (any, error) {
	whenEventWithoutTypes(input, func() {
		o.Types = o.defaultActivities()
	})
	return input, nil
}

func (o *OnWatch) defaultActivities() []EventWatchActivity {
	// all activity types
	return []EventWatchActivity{
		EventWatchActivityStarted,
	}
}

func (o *OnWorkflowRun) DecodeMapstructure(input any) (any, error) {
	whenEventWithoutTypes(input, func() {
		o.Types = o.defaultActivities()
	})
	return input, nil
}

func (o *OnWorkflowRun) defaultActivities() []EventWorkflowRunActivity {
	return []EventWorkflowRunActivity{
		EventWorkflowRunActivityCompleted,
		EventWorkflowRunActivityRequested,
	}
}

func whenEventWithoutTypes(input any, handler func()) {
	switch m := input.(type) {
	case map[string]any:
		v, ok := m["types"]
		if !ok || v == nil {
			handler()
			delete(m, "types")
		}
	}
}
