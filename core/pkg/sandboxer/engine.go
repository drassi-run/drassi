/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package sandboxer

import (
	"context"
	"io"

	"drassi.run/core/pkg/container"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/model/workflows"
)

type Engine interface {
	io.Closer

	Launch(context.Context, *LaunchRequest) (*LaunchResponse, error)
}

type LaunchRequest struct {
	Uid    string
	Github *records.Github

	JobContainer      *workflows.Container
	ServiceContainers map[string]*workflows.Container
}

type LaunchResponse struct {
	Sandbox         Sandbox
	ContainerEngine container.Engine

	JobContainer      *records.Container
	ServiceContainers map[string]*records.Container
}
