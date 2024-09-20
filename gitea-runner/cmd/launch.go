package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"code.gitea.io/actions-proto-go/runner/v1"
	"connectrpc.com/connect"
	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/executor/command"
	"drassi.run/core/pkg/executor/problem"
	"drassi.run/core/pkg/executor/reporter"
	"drassi.run/core/pkg/executor/secret"
	"drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/expression/libraries"
	"drassi.run/core/pkg/model"
	"drassi.run/core/pkg/model/dossiers"
	"drassi.run/core/pkg/model/workflows"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/sandboxer/host"
	"drassi.run/core/pkg/util/dig"
	"drassi.run/core/pkg/wire/cmdhandler"
	"drassi.run/core/pkg/wire/etc"
	"drassi.run/core/pkg/wire/reporter"
	"drassi.run/core/pkg/wire/streams"
	"drassi.run/gitea-runner/pkg/service"
	"github.com/spf13/cobra"
	"go.uber.org/dig"
	"golang.org/x/time/rate"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type launchCommand struct {
	runnerInfo RunnerInfo
	client     service.GiteaClient
	runtime    sandboxer.SandboxRuntime

	// tasksVersion used to store the version of the last task fetched from the Gitea.
	tasksVersion atomic.Int64
}

func NewLaunchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "launch",
		Short: "Start Gitea runner to receive request from server",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			command := launchCommand{}

			ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()
			go func() {
				<-ctx.Done()
				command.finalize(ctx)
			}()

			if err := command.initialize(ctx); err != nil {
				return err
			}
			return command.run(ctx)
		},
	}

	return cmd
}

func (c *launchCommand) initialize(ctx context.Context) error {
	if err := loadJson(".runner", &c.runnerInfo); err != nil {
		return err
	}

	c.client = service.NewClient(
		c.runnerInfo.Address,
		c.runnerInfo.InsecureSkipTLSVerify,
		c.runnerInfo.UUID,
		c.runnerInfo.Token,
	)

	req := &runnerv1.DeclareRequest{
		Version: "dev",
		Labels:  c.runnerInfo.Labels,
	}
	if _, err := c.client.Declare(ctx, connect.NewRequest(req)); err != nil {
		return err
	}

	c.runtime = getRuntime()
	return c.runtime.Connect(ctx)
}

func (c *launchCommand) run(ctx context.Context) error {
	// fetchInterval = 1s
	limiter := rate.NewLimiter(rate.Every(time.Second), 1)
	for {
		if err := limiter.Wait(ctx); err != nil {
			return err
		}

		task, ok := c.fetchTask(ctx)
		if !ok {
			continue
		}

		if err := c.runTask(ctx, task); err != nil {
			return err
		}
	}
}

func (c *launchCommand) fetchTask(ctx context.Context) (*runnerv1.Task, bool) {
	// fetchTimeout = 5s
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Load the version value that was in the cache when the request was sent.
	v := c.tasksVersion.Load()
	req := connect.NewRequest(&runnerv1.FetchTaskRequest{
		TasksVersion: v,
	})
	resp, err := c.client.FetchTask(ctx, req)
	if errors.Is(err, context.DeadlineExceeded) {
		err = nil
	}
	if err != nil {
		//log.WithError(err).Error("failed to fetch task")
		return nil, false
	}

	if resp == nil || resp.Msg == nil {
		return nil, false
	}

	if resp.Msg.TasksVersion > v {
		c.tasksVersion.CompareAndSwap(v, resp.Msg.TasksVersion)
	}

	if resp.Msg.Task == nil {
		return nil, false
	}

	// got a task, set `tasksVersion` to zero to force query db in the next request.
	c.tasksVersion.CompareAndSwap(resp.Msg.TasksVersion, 0)
	return resp.Msg.Task, true
}

func (c *launchCommand) runTask(ctx context.Context, task *runnerv1.Task) error {
	scope := dig.New().Scope("runner")

	// expression.Env
	needs := c.convertJobNeeds(task.Needs)
	opts := []expression.Option{
		expression.WithCache(true),
		expression.WithLibrary(libraries.StdLib()),
		expression.WithVariable("secrets", task.Secrets),
		expression.WithVariable("vars", task.Vars),
		expression.WithVariable("needs", needs),
		expression.WithAlias("gitea", "github"), // make `gitea` variable alias to `github`
	}
	if exprEnv, err := expression.NewEnv(opts...); err != nil {
		return err
	} else if err = xdig.Supply(scope, exprEnv); err != nil {
		return err
	}

	// Env context
	env := make(map[string]string)
	if err := xdig.Supply(scope, env); err != nil {
		return err
	}

	// Runner context
	runner := dossiers.Runner{
		Name:        c.runnerInfo.Name,
		Os:          model.Linux,
		Arch:        model.X64,
		Environment: "self-hosted",
	}
	if err := xdig.Supply(scope, runner); err != nil {
		return err
	}

	// GitHub context
	var github dossiers.Github
	if err := model.Decode(task.Context.AsMap(), &github); err != nil {
		return err
	} else if github.Token == "" {
		if t := task.Secrets["GITEA_TOKEN"]; t != "" {
			github.Token = t
		} else if t = task.Secrets["GITHUB_TOKEN"]; t != "" {
			github.Token = t
		}
	} else if err = xdig.Supply(scope, github); err != nil {
		return err
	}

	// secret.Masker
	sm := secret.NewMasker()
	for _, v := range task.Secrets {
		sm.AddSecret(secret.NewValueSecret(v))
	}
	if err := xdig.Supply(scope, sm); err != nil {
		return err
	}

	// problem.Matcher
	pm := make(map[string]problem.Matcher)
	if err := xdig.Supply(scope, pm); err != nil {
		return err
	}

	// Wire scope
	if err := scope.Provide(command.NewFileManager); err != nil {
		return err
	}
	if err := scope.Provide(newCommandConsoleManager); err != nil {
		return err
	}
	if err := scope.Provide(executor.NewSupervisor); err != nil {
		return err
	}
	if err := wire_cmdhandler.ProvideTo(scope); err != nil {
		return err
	}
	if err := wire_streams.ProvideTo(scope.Scope("internal(streams)")); err != nil {
		return err
	}
	if err := xdig.Supply(scope, c.runtime); err != nil {
		return err
	}
	if err := etc.Wire(scope); err != nil {
		return err
	}

	// JobRun
	workflow := new(workflows.Workflow)
	if err := c.decodeWorkflow(task.WorkflowPayload, workflow); err != nil {
		return err
	}
	jr, err := c.convertJobRun(workflow)
	if err != nil {
		return err
	}

	// reporter.Reporter
	rep := service.NewReporter(ctx, task.Id, jr, c.client)
	defer rep.Close()
	if err = xdig.Supply[reporter.Reporter](scope, rep); err != nil {
		return err
	} else if err = wire_reporter.ProvideTo(scope); err != nil {
		return err
	} else if err = wire_reporter.Wire(scope); err != nil {
		return err
	}

	scope = scope.Scope(fmt.Sprintf("job(%s)", jr.Id))
	je := executor.NewJobExecutor(jr)
	je.SetContext(ctx)

	defer je.Finalize()
	if err = je.Initialize(scope); err != nil {
		return err
	}

	r := je.RunJob()
	if r.Result != dossiers.ResultSuccess {
		return fmt.Errorf("")
	}

	return nil
}

type cmParams struct {
	dig.In
	Out io.Writer `name:"stdout"`
}

func newCommandConsoleManager(p cmParams) command.ConsoleManager {
	return command.NewConsoleManager(p.Out)
}

func (c *launchCommand) convertJobRun(wf *workflows.Workflow) (*executor.JobRun, error) {
	if len(wf.Jobs) > 1 {
		return nil, errors.New("multiple jobs found")
	}
	for jobId, job := range wf.Jobs {
		if nj, ok := job.(*workflows.NormalJob); ok {
			jr := executor.ToJobRun(jobId, nj)
			return jr, nil
		}
		return nil, fmt.Errorf("unsupported job type %T", job)
	}
	return nil, fmt.Errorf("empty job")
}

func (c *launchCommand) convertJobNeeds(taskNeeds map[string]*runnerv1.TaskNeed) map[string]*dossiers.Need {
	if len(taskNeeds) == 0 {
		return nil
	}

	needs := make(map[string]*dossiers.Need, len(taskNeeds))
	for k, n := range taskNeeds {
		needs[k] = &dossiers.Need{
			Outputs: n.Outputs,
			Result:  resultMap[n.Result],
		}
	}
	return needs
}

func (c *launchCommand) decodeWorkflow(payload []byte, workflow *workflows.Workflow) error {
	var raw any
	reader := bytes.NewReader(payload)
	if err := yaml.NewDecoder(reader).Decode(&raw); err != nil && err != io.EOF {
		return err
	}

	return model.Decode(raw, workflow)
}

func (c *launchCommand) finalize(ctx context.Context) {
}

func getRuntime() sandboxer.SandboxRuntime {
	//config := &incus.Incus{
	//	TypeMeta: metav1.TypeMeta{
	//		APIVersion: "sandboxer.drasi.run/v1",
	//		Kind:       "Incus",
	//	},
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name: "default",
	//	},
	//	Spec: incus.IncusSpec{
	//		Endpoint: "unix://",
	//		Template: incus.IncusTemplateSpec{
	//			Source: api.InstanceSource{
	//				Type:     "image",
	//				Alias:    "ubuntu/22.04",
	//				Server:   "https://images.linuxcontainers.org",
	//				Protocol: "simplestreams",
	//				//Mode:     "pull",
	//			},
	//			Type: "container",
	//		},
	//	},
	//}
	//return incus.NewSandboxRuntime(config)

	config := &host.Host{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "sandboxer.drasi.run/v1",
			Kind:       "Host",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "default",
		},
		Spec: host.HostSpec{
			RootDir: "/tmp/gitea-runner",
		},
	}
	return host.NewSandboxRuntime(config)
}

func loadJson(file string, object any) error {
	f, err := os.OpenFile(file, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return err
	}
	return json.NewDecoder(f).Decode(object)
}

var resultMap = map[runnerv1.Result]dossiers.Result{
	runnerv1.Result_RESULT_UNSPECIFIED: "",
	runnerv1.Result_RESULT_SUCCESS:     dossiers.ResultSuccess,
	runnerv1.Result_RESULT_FAILURE:     dossiers.ResultFailure,
	runnerv1.Result_RESULT_CANCELLED:   dossiers.ResultCancelled,
	runnerv1.Result_RESULT_SKIPPED:     dossiers.ResultSkipped,
}
