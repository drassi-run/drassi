/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package sandboxer

import (
	"bytes"
	"fmt"
	"strings"
	"sync"

	"drassi.run/core/config"
	"drassi.run/core/pkg/store/oci"
	"github.com/pelletier/go-toml/v2"
	"github.com/pelletier/go-toml/v2/unstable"
)

type Factory interface {
	// SupportContainer(config) // TODO

	SetOciStore(store ocistore.Manager)
	ProvisionRuntime(config map[string]*config.Runtime)
	Create() (Engine, error)
}

var (
	mu        sync.RWMutex
	defaults  = make(map[string]func() any)
	factories = make(map[string]func(cfg unstable.RawMessage) (Factory, error))
)

func Register[T any](provider string, d func() T, fn func(cfg T) Factory) {
	provider = strings.ToLower(provider)
	mu.Lock()
	defer mu.Unlock()

	defaults[provider] = func() any {
		return d()
	}
	factories[provider] = func(raw unstable.RawMessage) (Factory, error) {
		cfg := d()
		if len(raw) > 0 {
			if err := toml.NewDecoder(bytes.NewReader(raw)).EnableUnmarshalerInterface().Decode(cfg); err != nil {
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
