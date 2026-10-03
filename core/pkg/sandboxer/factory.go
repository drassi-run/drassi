/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package sandboxer

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"drassi.run/core/config"
	"drassi.run/core/pkg/store/oci"
)

type Factory interface {
	// SupportContainer(config) // TODO

	RootDir(string)
	SetOciStore(ocistore.Manager)
	ProvisionRuntime(map[string]*config.Runtime)

	Create() (Engine, error)
}

var (
	mu        sync.RWMutex
	defaults  = make(map[string]func() any)
	factories = make(map[string]func(cfg json.RawMessage) (Factory, error))
)

func Register[T any](provider string, d func() T, fn func(cfg T) Factory) {
	provider = strings.ToLower(provider)
	mu.Lock()
	defer mu.Unlock()

	defaults[provider] = func() any {
		return d()
	}
	factories[provider] = func(raw json.RawMessage) (Factory, error) {
		cfg := d()
		if len(raw) > 0 {
			if err := json.Unmarshal(raw, cfg); err != nil {
				return nil, fmt.Errorf("unmarshal provider %q config: %v", provider, err)
			}
		}

		return fn(cfg), nil
	}
}

func NewFactory(config *config.Sandboxer) (Factory, error) {
	provider := strings.ToLower(config.Provider)
	mu.RLock()
	fn, ok := factories[provider]
	mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unsupported sandboxer provider %q", config.Provider)
	}
	return fn(config.Config)
}

func DefaultConfig(provider string) any {
	provider = strings.ToLower(provider)
	mu.RLock()
	fn, ok := defaults[provider]
	mu.RUnlock()

	if !ok {
		return nil
	}
	return fn()
}

func SupportedProviders() []string {
	mu.RLock()
	defer mu.RUnlock()

	providers := make([]string, 0, len(defaults))
	for k := range defaults {
		providers = append(providers, k)
	}
	return providers
}
