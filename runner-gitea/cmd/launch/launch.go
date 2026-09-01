/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package launch

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"connectrpc.com/connect"
	coreconfig "drassi.run/core/config"
	"drassi.run/core/pkg/model"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/store/repository/gitstore"
	"drassi.run/core/util/dig"
	"drassi.run/core/wire"
	giteaconfig "drassi.run/gitea-runner/config"
	"drassi.run/gitea-runner/pkg/gitea"
	"drassi.run/gitea-runner/pkg/worker"
	"gitea.dev/actionslib/runner/v1"
	"github.com/chainguard-dev/clog"
	"github.com/spf13/cobra"
	"go.uber.org/dig"
	"golang.org/x/time/rate"
)

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
	var opts options

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

func (c *launcher) Init(ctx context.Context, o *options) error {
	clog.InfoContextf(ctx, "initializing gitea-runner")

	config, err := LoadConfig(o)
	if err != nil {
		return err
	}

	spec := config.Runner
	c.runnerName = spec.Name
	c.concurrency = spec.Concurrency
	if c.concurrency == 0 {
		c.concurrency = 5
	}
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
	return c.loadSandboxer(config, config.UseSandboxer)
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

func (c *launcher) runTaskE(ctx context.Context, task *runnerv1.Task) error {
	w := worker.New(task)
	return w.Run(ctx, c.module())
}

func (c *launcher) loadGitStore() error {
	if store, err := gitstore.New(".cache"); err != nil {
		return err
	} else {
		c.store = store
	}
	return nil
}

func (c *launcher) loadSandboxer(config *giteaconfig.Config, name string) error {
	if sbConfig, ok := config.Sandboxers[name]; !ok {
		return fmt.Errorf("sandboxer %q not configured", name)
	} else if engine, err := coreconfig.NewSandboxerEngine(sbConfig); err != nil {
		return err
	} else {
		c.runtime = engine
		return nil
	}
}

func (c *launcher) module() *wire.Module {
	fn := func(scope *dig.Scope) error {
		runner := &records.RunnerInfo{
			Name:        c.runnerName,
			Os:          model.Linux,
			Arch:        model.X64,
			Environment: "self-hosted",
		}
		if err := xdig.Supply(scope, runner); err != nil {
			return fmt.Errorf("provide records.Runner: %w", err)
		}
		if err := xdig.Supply(scope, c.runtime); err != nil {
			return fmt.Errorf("provide sandboxer.Engine: %w", err)
		}
		if err := xdig.Supply(scope, c.store); err != nil {
			return fmt.Errorf("provide gitstore.Store: %w", err)
		}
		if err := xdig.Supply(scope, c.client); err != nil {
			return fmt.Errorf("provide gitea.Client: %w", err)
		}
		return nil
	}
	return wire.NewModule("gitea/launch", fn)
}
