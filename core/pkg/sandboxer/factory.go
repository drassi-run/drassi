/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package sandboxer

import (
	"fmt"
	"strings"
	"sync"

	"drassi.run/core/config"
	"github.com/pelletier/go-toml/v2/unstable"
)

type Factory func(cfg unstable.RawMessage) (Engine, error)

var (
	mu        sync.RWMutex
	factories = make(map[string]Factory)
)

func Register(provider string, factory Factory) {
	mu.Lock()
	defer mu.Unlock()
	factories[strings.ToLower(provider)] = factory
}

func NewEngine(config *config.Sandboxer) (Engine, error) {
	provider := strings.ToLower(config.Provider)
	mu.RLock()
	factory, ok := factories[provider]
	mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unsupported sandboxer provider %q", config.Provider)
	}
	return factory(config.Config)
}
