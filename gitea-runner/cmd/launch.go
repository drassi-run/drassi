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

	runnerv1 "code.gitea.io/actions-proto-go/runner/v1"
	"connectrpc.com/connect"
	"github.com/dungdm93/drassi/core/pkg/executor"
	"github.com/dungdm93/drassi/core/pkg/model"
	"github.com/dungdm93/drassi/core/pkg/model/workflows"
	"github.com/dungdm93/drassi/core/pkg/sandboxer"
	"github.com/dungdm93/drassi/core/pkg/sandboxer/incus"
	"github.com/dungdm93/drassi/gitea-runner/pkg/service"
	"github.com/lxc/incus/shared/api"
	"github.com/spf13/cobra"
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

	c.runtime = incus.NewSandboxRuntime(newDefaultConfig())
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

	// got a task, set `tasksVersion` to zero to force query db in next request.
	c.tasksVersion.CompareAndSwap(resp.Msg.TasksVersion, 0)
	return resp.Msg.Task, true
}

func (c *launchCommand) runTask(ctx context.Context, task *runnerv1.Task) error {
	workflow, err := c.decodeWorkflow(task.WorkflowPayload)
	if err != nil {
		return err
	}

	jr, err := c.convertJobRun(workflow)
	if err != nil {
		return err
	}

	reporter := service.NewReporter(ctx, task.Id, jr, c.client)
	defer reporter.Close()

	je := executor.JobExecutor{
		JobRun:   jr,
		Reporter: reporter,
	}

	if err = je.Initialize(ctx, c.runtime); err != nil {
		return err
	}

	if err = je.RunJob(ctx); err != nil {
		return err
	}

	return je.Finalize(ctx, c.runtime)
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

func (c *launchCommand) decodeWorkflow(payload []byte) (*workflows.Workflow, error) {
	var raw any
	reader := bytes.NewReader(payload)
	if err := yaml.NewDecoder(reader).Decode(&raw); err != nil && err != io.EOF {
		return nil, err
	}

	workflow := new(workflows.Workflow)
	if err := model.Decode(raw, workflow); err != nil {
		return nil, err
	}
	return workflow, nil
}

func (c *launchCommand) finalize(ctx context.Context) {
}

func newDefaultConfig() *incus.Incus {
	return &incus.Incus{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "sandboxer.drasi.run/v1",
			Kind:       "Incus",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "default",
		},
		Spec: incus.IncusSpec{
			Endpoint: "unix://",
			Template: incus.IncusTemplateSpec{
				Source: api.InstanceSource{
					Type:     "image",
					Alias:    "ubuntu/22.04",
					Server:   "https://images.linuxcontainers.org",
					Protocol: "simplestreams",
					//Mode:     "pull",
				},
				Type: "container",
			},
		},
	}
}

func loadJson(file string, object any) error {
	f, err := os.OpenFile(file, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return err
	}
	return json.NewDecoder(f).Decode(object)
}
