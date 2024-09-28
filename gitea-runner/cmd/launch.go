package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"code.gitea.io/actions-proto-go/runner/v1"
	"connectrpc.com/connect"
	"drassi.run/core/pkg/model"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/sandboxer/host"
	"drassi.run/core/pkg/store/repository/gitstore"
	"drassi.run/core/pkg/util/dig"
	"drassi.run/gitea-runner/pkg/service"
	"drassi.run/gitea-runner/pkg/worker"
	"github.com/spf13/cobra"
	"go.uber.org/dig"
	"golang.org/x/time/rate"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type launchCommand struct {
	runnerInfo RunnerInfo
	client     service.GiteaClient
	runtime    sandboxer.SandboxRuntime
	store      gitstore.Store

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

	if store, err := gitstore.New(".cache"); err != nil {
		return err
	} else {
		c.store = store
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

	// Runner context
	runner := records.Runner{
		Name:        c.runnerInfo.Name,
		Os:          model.Linux,
		Arch:        model.X64,
		Environment: "self-hosted",
	}
	if err := xdig.Supply(scope, runner); err != nil {
		return err
	}
	if err := xdig.Supply(scope, c.runtime); err != nil {
		return err
	}
	if err := xdig.Supply(scope, c.store); err != nil {
		return err
	}
	if err := xdig.Supply(scope, c.client); err != nil {
		return err
	}

	w := worker.New(ctx, task)
	if err := w.Setup(scope); err != nil {
		return err
	}
	defer func(w *worker.Worker) {
		if err := w.Teardown(); err != nil {
			fmt.Printf("Error while teardown worker: %v\n", err)
		}
	}(w)
	return w.Run()
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
