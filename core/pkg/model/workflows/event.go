/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package workflows

import "encoding/json/jsontext"

// The name of the GitHub event that triggers the workflow.
// You can provide a single event string, array of events, array of event types, or an event configuration map
// that schedules a workflow or restricts the execution of a workflow to specific files, tags, or branch changes.
// For a list of available events, see https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows.
// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#on
type On map[string]Event

// Event https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows
type Event = jsontext.Value
