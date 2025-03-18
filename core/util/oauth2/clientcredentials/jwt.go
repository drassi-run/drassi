/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Cloned from https://github.com/golang/oauth2/pull/450 with some refactor
package clientcredentials

import (
	"math/rand"
	"time"

	"golang.org/x/oauth2/jws" //nolint:staticcheck // SA1019 - jws for internal used
)

const (
	clientAssertionType = "urn:ietf:params:oauth:client-assertion-type:jwt-bearer"
	letters             = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
)

func randJWTID(n int) string {
	b := make([]byte, n)
	l := int64(len(letters))
	for i := range b {
		b[i] = letters[rand.Int63n(l)]
	}
	return string(b)
}

func (c *tokenSource) jwtAssertion() (string, error) {
	now := time.Now()
	claimSet := &jws.ClaimSet{
		Iss: c.conf.ClientID,
		Sub: c.conf.ClientID,
		Aud: c.conf.TokenURL,

		PrivateClaims: map[string]any{
			"jti": randJWTID(36),
			"nbf": now.Unix(),
		},
	}
	if t := c.conf.JWTExpires; t > 0 {
		claimSet.Exp = now.Add(t).Unix()
	} else {
		claimSet.Exp = now.Add(time.Hour).Unix()
	}

	h := &jws.Header{
		Algorithm: "RS256",
		Typ:       "JWT",
		KeyID:     c.conf.KeyID,
	}
	return jws.Encode(h, claimSet, c.conf.PrivateKey)
}
