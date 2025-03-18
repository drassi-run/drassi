/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

// +k8s:deepcopy-gen=package
// +k8s:defaulter-gen=TypeMeta

// Define command alias
//go:generate -command deepcopy-gen go tool deepcopy-gen
//go:generate -command defaulter-gen go tool defaulter-gen

//go:generate deepcopy-gen --output-file zz_generated.deepcopy.go drassi.run/gitea-runner/pkg/apis/v1alpha1
//go:generate defaulter-gen --output-file zz_generated.defaulter.go drassi.run/gitea-runner/pkg/apis/v1alpha1
package v1alpha1
