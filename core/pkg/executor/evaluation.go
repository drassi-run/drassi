package executor

import (
	"context"

	"drassi.run/core/pkg/model/dossiers"
)

type evaluationSupplier struct {
	dossier *dossiers.Dossier
}

func (e *evaluationSupplier) Values(name string) context.Context {
	return context.Background()
}

func (e *evaluationSupplier) Functions(name string) []string {
	return nil
}

func (e *evaluationSupplier) DefaultValue(name string) any {
	return nil
}
