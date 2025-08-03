/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package support

import "maps"

type EnvProvider interface {
	Env() map[string]string

	ProvideEnv(func() map[string]string)
}

func StaticEnv(m map[string]string) func() map[string]string {
	return func() map[string]string {
		return m
	}
}

func CIEnv() func() map[string]string {
	return func() map[string]string {
		return map[string]string{
			"CI":             "true",
			"GITHUB_ACTIONS": "true",
		}
	}
}

func NewEnvProvider() EnvProvider {
	return new(envProvider)
}

type envProvider []func() map[string]string

func (ep *envProvider) Env() map[string]string {
	m := make(map[string]string)
	for _, fn := range *ep {
		maps.Copy(m, fn())
	}
	return m
}

func (ep *envProvider) ProvideEnv(f func() map[string]string) {
	*ep = append(*ep, f)
}
