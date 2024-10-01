package types

import "time"

type Cache struct {
	ID         uint64    `json:"id"` // generated
	Key        string    `json:"key"`
	Version    string    `json:"version"`
	Namespace  string    `json:"namespace"`
	Size       int64     `json:"cacheSize"`
	Complete   bool      `json:"complete"`
	CreatedAt  time.Time `json:"createdAt"`
	LastUsedAt time.Time `json:"lastUsedAt"`
}
