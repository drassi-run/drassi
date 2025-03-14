package scribe

import "context"

type Output interface {
	SetDebug(b bool)
	EnableDebug() bool
	Inscribe(ctx context.Context, message string) error
}

type discard struct{}

func (h discard) SetDebug(bool)                          {}
func (h discard) EnableDebug() bool                      { return false }
func (h discard) Inscribe(context.Context, string) error { return nil }
