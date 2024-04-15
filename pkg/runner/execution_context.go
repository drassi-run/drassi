package runner

// TODO: implement

type IExecutionContext interface {
	NameOf() string
	JobContext() *JobContext
	IsEmbedded() bool
	Stage() ActionRunStage
	GetGitHubContext(name string) string
}
type JobContext struct {
	Status ActionResult
}

type ActionResult int

const (
	ActionResultSuccess ActionResult = iota
	ActionResultFailure
	ActionResultCancelled
	ActionResultSkipped
)

func (a ActionResult) String() string {
	switch a {
	case ActionResultSuccess:
		return "Success"
	case ActionResultFailure:
		return "Failure"
	case ActionResultCancelled:
		return "Cancelled"
	case ActionResultSkipped:
		return "Skipped"
	default:
		return "Unknown ActionResult"
	}
}

var AvailableActionResults = map[ActionResult]struct{}{
	ActionResultSuccess:   {},
	ActionResultFailure:   {},
	ActionResultCancelled: {},
	ActionResultSkipped:   {},
}

func TryParseActionResult(unknown string) (ok bool) {
	for a, _ := range AvailableActionResults {
		if unknown == a.String() {
			return true
		}
	}
	return false
}

type ActionRunStage int

const (
	ActionRunStagePre ActionRunStage = iota
	ActionRunStageMain
	ActionRunStagePost
)

func (a ActionRunStage) String() string {
	switch a {
	case ActionRunStagePre:
		return "pre"
	case ActionRunStageMain:
		return "main"
	case ActionRunStagePost:
		return "post"
	default:
		return "ActionRunStage ActionResult"
	}
}
