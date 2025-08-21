/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package launch

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"code.gitea.io/actions-proto-go/runner/v1"
	"connectrpc.com/connect"
	"drassi.run/core/pkg/manifest"
	"drassi.run/core/pkg/model"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/sandboxer"
	sandboxerv1a1 "drassi.run/core/pkg/sandboxer/apis/v1alpha1"
	"drassi.run/core/pkg/sandboxer/container"
	"drassi.run/core/pkg/sandboxer/host"
	"drassi.run/core/pkg/sandboxer/incus"
	"drassi.run/core/pkg/store/repository/gitstore"
	"drassi.run/core/util/dig"
	"drassi.run/core/util/error"
	"drassi.run/gitea-runner/cmd/common"
	giteav1a1 "drassi.run/gitea-runner/pkg/apis/v1alpha1"
	"drassi.run/gitea-runner/pkg/gitea"
	"drassi.run/gitea-runner/pkg/worker"
	"github.com/chainguard-dev/clog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.uber.org/dig"
	"golang.org/x/time/rate"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type option struct {
	common.Options

	name string
}

func (o *option) RegisterFlags(flags *pflag.FlagSet) {
	o.Options.RegisterFlags(flags)

	flags.StringVar(&o.name, "name", "", "Gitea instance name")
	_ = cobra.MarkFlagRequired(flags, "name")
}

type launcher struct {
	runnerName  string
	concurrency int
	client      gitea.Client
	runtime     sandboxer.Engine
	store       gitstore.Store

	// tasksVersion used to store the version of the last task fetched from the Gitea.
	tasksVersion atomic.Int64
}

func New() *cobra.Command {
	var opts option

	cmd := &cobra.Command{
		Use:   "launch",
		Short: "Start Gitea runner to receive request from server",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			l := new(launcher)

			if err := l.Init(ctx, &opts); err != nil {
				return err
			}
			return l.Run(ctx)
		},
	}

	flags := cmd.Flags()
	opts.RegisterFlags(flags)

	return cmd
}

func (c *launcher) Init(ctx context.Context, opts *option) error {
	clog.InfoContextf(ctx, "initializing gitea-runner")

	store, err := common.ManifestStore(&opts.Options)
	if err != nil {
		return err
	}

	o, err := c.loadGiteaManifest(ctx, store, opts.name)
	if err != nil {
		return err
	}

	spec := o.Spec
	c.concurrency = spec.Concurrency
	c.client = gitea.NewClient(
		spec.Address, spec.InsecureSkipTLSVerify,
		spec.UUID, spec.Token,
	)

	req := &runnerv1.DeclareRequest{
		Version: "dev",
		Labels:  spec.RunnerLabels,
	}
	if _, err = c.client.Declare(ctx, connect.NewRequest(req)); err != nil {
		return err
	}

	if err = c.loadGitStore(); err != nil {
		return err
	}
	return c.loadSandboxer(ctx, store, spec.SandboxerRef)
}

func (c *launcher) Run(ctx context.Context) error {
	clog.InfoContextf(ctx, "gitea-runner started")

	// fetchInterval = 1s
	limiter := rate.NewLimiter(rate.Every(time.Second), 1)

	availableCh := make(chan struct{}, c.concurrency)
	workerPool := worker.NewPool(c.concurrency, availableCh, c.runTask)

	if err := workerPool.Start(ctx); err != nil {
		return err
	}

	for {
		if err := limiter.Wait(ctx); err != nil {
			return err
		}
		<-availableCh

		task, ok := c.fetchTask(ctx)
		if !ok {
			availableCh <- struct{}{}
			continue
		}

		workerPool.Submit(task)
	}
}

func (c *launcher) fetchTask(ctx context.Context) (*runnerv1.Task, bool) {
	clog.DebugContextf(ctx, "Fetching task")

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
		clog.ErrorContextf(ctx, "failed to fetch task: %v", err)
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

func (c *launcher) runTask(ctx context.Context, task *runnerv1.Task) {
	err := c.runTaskE(ctx, task)
	if err != nil {
		clog.ErrorContextf(ctx, "Failed to run task: %v", err)
	}
}

func (c *launcher) runTaskE(ctx context.Context, task *runnerv1.Task) (err error) {
	defer xerror.Recover(&err)

	scope := dig.New().Scope("runner")

	// Runner context
	runner := records.Runner{
		Name:        c.runnerName,
		Os:          model.Linux,
		Arch:        model.X64,
		Environment: "self-hosted",
	}
	if err = xdig.Supply(scope, runner); err != nil {
		return
	}
	if err = xdig.Supply(scope, c.runtime); err != nil {
		return
	}
	if err = xdig.Supply(scope, c.store); err != nil {
		return
	}
	if err = xdig.Supply(scope, c.client); err != nil {
		return
	}

	w := worker.New(task)
	return w.Run(ctx, scope)
}

func (c *launcher) loadGiteaManifest(ctx context.Context, store manifest.Store, name string) (*giteav1a1.GiteaRunner, error) {
	gvk := giteav1a1.SchemeGroupVersion.WithKind("GiteaRunner")

	if o, err := store.Load(ctx, gvk, name); err != nil {
		return nil, err
	} else {
		return o.(*giteav1a1.GiteaRunner), nil
	}
}

func (c *launcher) loadGitStore() error {
	if store, err := gitstore.New(".cache"); err != nil {
		return err
	} else {
		c.store = store
	}
	return nil
}

func (c *launcher) loadSandboxer(ctx context.Context, store manifest.Store, ref corev1.TypedLocalObjectReference) (err error) {
	gv := sandboxerv1a1.SchemeGroupVersion
	if ref.APIGroup != nil {
		if gv, err = schema.ParseGroupVersion(*ref.APIGroup); err != nil {
			return err
		}
	}

	gvk := gv.WithKind(ref.Kind)
	o, err := store.Load(ctx, gvk, ref.Name)
	if err != nil {
		return err
	}

	switch o := o.(type) {
	case *sandboxerv1a1.ContainerSandboxer:
		c.runtime, err = container.New(&o.Spec)
	case *sandboxerv1a1.HostSandboxer:
		c.runtime, err = host.New(&o.Spec)
	case *sandboxerv1a1.IncusSandboxer:
		c.runtime, err = incus.New(&o.Spec)
	}

	c.runtime = sandboxer.WithTelemetry(c.runtime)
	return
}
