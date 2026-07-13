/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package launch

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	ghaconfig "drassi.run/gha-runner/config"
	"github.com/pelletier/go-toml/v2"
	"golang.org/x/oauth2"
)

func loadConfig(path string) (*ghaconfig.Config, error) {
	if path == "" {
		return nil, fmt.Errorf("--config is required")
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	config := ghaconfig.DefaultConfig()
	dec := toml.NewDecoder(f).EnableUnmarshalerInterface()
	if err = dec.Decode(config); err != nil {
		return nil, err
	}
	return config, nil
}

func fixupToken(token *oauth2.Token) error {
	if token != nil && strings.EqualFold(token.TokenType, "jwt") {
		token.TokenType = "Bearer"
	}
	return nil
}

func decodeKey(auth ghaconfig.RunnerAuthorization, configFile string) (*rsa.PrivateKey, error) {
	var keyData []byte
	if data := auth.PrivateKey; data != "" {
		if b, err := base64.StdEncoding.DecodeString(data); err == nil {
			keyData = b
		} else {
			return nil, fmt.Errorf("invalid private_key base64 std format: %v", err)
		}
	} else if file := auth.PrivateKeyFile; file != "" {
		if !filepath.IsAbs(file) {
			configDir := filepath.Dir(configFile)
			file = filepath.Join(configDir, file)
		}
		if b, err := os.ReadFile(file); err != nil {
			return nil, fmt.Errorf("read private_key_file: %v", err)
		} else {
			keyData = b
		}
	} else {
		return nil, fmt.Errorf("authorization.private_key or authorization.private_key_file is required")
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("decode PEM block error")
	}

	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		if privateKey, ok := key.(*rsa.PrivateKey); ok {
			return privateKey, nil
		}
	}

	return nil, fmt.Errorf("parse RSA private_key error")
}
