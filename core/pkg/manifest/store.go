package manifest

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Store interface {
	Load(ctx context.Context, gvk schema.GroupVersionKind, name string) (runtime.Object, error)
}
