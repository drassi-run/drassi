/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

// +groupName=sandboxer.drassi.run
// +k8s:deepcopy-gen=package

// Define command alias
//go:generate -command deepcopy-gen go tool deepcopy-gen
//go:generate -command register-gen go tool register-gen

//go:generate deepcopy-gen --output-file zz_generated.deepcopy.go drassi.run/core/pkg/sandboxer/apis/v1alpha1
//go:generate register-gen --output-file zz_generated.register.go drassi.run/core/pkg/sandboxer/apis/v1alpha1
package v1alpha1
