/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type GitHubRunner struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec GitHubRunnerSpec `json:"spec,omitempty"`
}

type GitHubRunnerSpec struct {
	RunnerId        int    `json:"runnerId"`   // a.k.a agentId
	RunnerName      string `json:"runnerName"` // a.k.a agentName
	GroupId         int    `json:"groupId"`    // a.k.a poolId
	GroupName       string `json:"groupName"`  // a.k.a poolName
	RegistrationURL string `json:"registrationUrl"`

	Authorization GitHubRunnerAuthorization `json:"authorization"`
}

type GitHubRunnerAuthorization struct {
	Url       string                      `json:"url"`
	ClientId  string                      `json:"clientId"`
	SecretRef corev1.LocalObjectReference `json:"secretRef"` // reference to the public key secret
}
