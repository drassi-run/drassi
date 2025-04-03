/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package launch

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"path/filepath"
	"strings"

	"drassi.run/core/pkg/manifest"
	"drassi.run/core/pkg/manifest/filesystem"
	"drassi.run/core/pkg/sandboxer"
	sandboxerv1a1 "drassi.run/core/pkg/sandboxer/apis/v1alpha1"
	"drassi.run/core/pkg/sandboxer/container"
	"drassi.run/core/pkg/sandboxer/host"
	"drassi.run/core/pkg/sandboxer/incus"
	ghav1a1 "drassi.run/gha-runner/pkg/apis/v1alpha1"
	giteav1a1 "drassi.run/gitea-runner/pkg/apis/v1alpha1"
	"golang.org/x/oauth2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
)

func manifestStore(o *options) (manifest.Store, error) {
	s, err := newScheme()
	if err != nil {
		return nil, err
	}

	if o.Store == "local" {
		if o.ConfigDir == "" {
			return nil, fmt.Errorf("--config-dir is required")
		}

		absPath, err := filepath.Abs(o.ConfigDir)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve config-dir path: %v", err)
		}
		store := filesystem.NewStore(absPath, s)
		return store, nil
	}

	return nil, fmt.Errorf("unknown manifest store: %s", o.Store)
}

func newScheme() (*runtime.Scheme, error) {
	s := runtime.NewScheme()

	if err := scheme.AddToScheme(s); err != nil {
		return nil, err
	}
	if err := sandboxerv1a1.AddToScheme(s); err != nil {
		return nil, err
	}
	if err := ghav1a1.AddToScheme(s); err != nil {
		return nil, err
	}
	return s, nil
}

func loadRunnerManifest(ctx context.Context, store manifest.Store, name string) (*ghav1a1.GitHubRunner, error) {
	gvk := giteav1a1.SchemeGroupVersion.WithKind("GitHubRunner")

	if o, err := store.Load(ctx, gvk, name); err != nil {
		return nil, err
	} else {
		return o.(*ghav1a1.GitHubRunner), nil
	}
}

func loadSecretManifest(ctx context.Context, store manifest.Store, name string) (*corev1.Secret, error) {
	gvk := corev1.SchemeGroupVersion.WithKind("Secret")

	if o, err := store.Load(ctx, gvk, name); err != nil {
		return nil, err
	} else {
		return o.(*corev1.Secret), nil
	}
}

func loadSandboxer(ctx context.Context, store manifest.Store, ref corev1.TypedLocalObjectReference) (s sandboxer.Engine, err error) {
	gv := sandboxerv1a1.SchemeGroupVersion
	if ref.APIGroup != nil {
		if gv, err = schema.ParseGroupVersion(*ref.APIGroup); err != nil {
			return nil, err
		}
	}

	gvk := gv.WithKind(ref.Kind)
	o, err := store.Load(ctx, gvk, ref.Name)
	if err != nil {
		return nil, err
	}

	switch o := o.(type) {
	case *sandboxerv1a1.ContainerSandboxer:
		s, err = container.New(&o.Spec)
	case *sandboxerv1a1.HostSandboxer:
		s, err = host.New(&o.Spec)
	case *sandboxerv1a1.IncusSandboxer:
		s, err = incus.New(&o.Spec)
	}

	if s != nil {
		s = sandboxer.WithTelemetry(s)
	}
	return
}

func fixupToken(token *oauth2.Token) error {
	if token != nil && strings.EqualFold(token.TokenType, "jwt") {
		token.TokenType = "Bearer"
	}
	return nil
}

func decodeKey(secret *corev1.Secret) (*rsa.PrivateKey, error) {
	var k []byte
	if b, ok := secret.Data["private.key"]; ok {
		enc := base64.StdEncoding
		k = make([]byte, enc.DecodedLen(len(b)))
		if _, err := enc.Decode(k, b); err != nil {
			return nil, err
		}
	} else if s, ok := secret.StringData["private.key"]; ok {
		k = []byte(s)
	} else {
		return nil, fmt.Errorf("no private.key found in secret")
	}

	block, _ := pem.Decode(k)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		return key.(*rsa.PrivateKey), nil
	}

	return nil, fmt.Errorf("failed to parse RSA private.key")
}
