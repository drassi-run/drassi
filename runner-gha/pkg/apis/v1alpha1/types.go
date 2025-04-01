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

	Spec   GitHubRunnerSpec   `json:"spec,omitempty"`
	Status GitHubRunnerStatus `json:"status,omitempty"`
}

type GitHubRunnerSpec struct {
	RunnerId        int32  `json:"runnerId"` // a.k.a agentId
	GroupId         int32  `json:"groupId"`  // a.k.a poolId
	ServerUrl       string `json:"serverUrl"`
	RegistrationUrl string `json:"registrationUrl"`

	Authorization GitHubRunnerAuthorization `json:"authorization,omitempty"`
}

type GitHubRunnerAuthorization struct {
	Url       string                      `json:"url"`
	ClientId  string                      `json:"clientId"`
	SecretRef corev1.LocalObjectReference `json:"secretRef"` // reference to the public key secret
}

type GitHubRunnerStatus struct {
	RunnerName string   `json:"runnerName,omitempty"` // a.k.a agentName
	GroupName  string   `json:"groupName,omitempty"`  // a.k.a poolName
	Labels     []string `json:"labels,omitempty"`
}
