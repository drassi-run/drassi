/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package cmdtypes

import (
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/runtime"
)

type SupportAddPath interface {
	AddPath(paths []string)
}

type SupportSetEnv interface {
	SetEnv(env map[string]string)
}

type SupportSetOutput interface {
	SetOutput(output map[string]string)
}

type SupportSaveState interface {
	SaveState(state map[string]string)
}

type HasPathTranslator interface {
	PathTranslator() runtime.PathTranslator
}

type HasForge interface {
	Forge() *records.Forge
}
