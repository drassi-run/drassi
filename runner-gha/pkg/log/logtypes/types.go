/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package logtypes

import "drassi.run/gha-runner/pkg/log"

type Subscriber interface {
	Run(ch <-chan *log.Event)
	Wait()
}

type SignedUrlResponse interface {
	GetUrl() string
	GetStorageType() string
}

const StorageAzureBlob = "BLOB_STORAGE_TYPE_AZURE"

type Stat struct {
	Lines int
	Size  int64
}

func NewStat(lines int, size int64) *Stat {
	return &Stat{lines, size}
}
