package filesystem

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"drassi.run/core/pkg/manifest"
	"drassi.run/core/util/io"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
)

var metadataAccessor = meta.NewAccessor()

type key struct {
	schema.GroupVersionKind
	Name string
}

type store struct {
	dir     string
	fsys    billy.Filesystem
	scheme  *runtime.Scheme
	decoder runtime.Decoder

	once  sync.Once
	cache map[key]runtime.Object
	err   error
}

func NewStore(dir string, scheme *runtime.Scheme) manifest.Store {
	s := &store{
		dir:     dir,
		fsys:    osfs.New("/"),
		scheme:  scheme,
		decoder: unstructured.UnstructuredJSONScheme,
		cache:   make(map[key]runtime.Object),
	}
	return s
}

// Load retrieves an object by its GroupVersionKind and name
func (s *store) Load(ctx context.Context, gvk schema.GroupVersionKind, name string) (runtime.Object, error) {
	s.once.Do(func() { s.sync(ctx) })
	if err := s.err; err != nil {
		return nil, err
	}

	k := key{GroupVersionKind: gvk, Name: name}
	if obj, exists := s.cache[k]; exists {
		return obj, nil
	}

	return nil, errors.New("object not found")
}

func (s *store) sync(ctx context.Context) {
	s.err = s.syncErr(ctx)
}

func (s *store) syncErr(ctx context.Context) error {
	entries, err := s.fsys.ReadDir(s.dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if err = ctx.Err(); err != nil {
			return err
		}

		name := entry.Name()
		path := filepath.Join(s.dir, name)
		if strings.HasPrefix(name, ".") {
			log.Printf("Skip hidden file %q", path)
			continue
		}

		switch mode := entry.Mode(); {
		case mode.IsRegular() || mode&os.ModeSymlink != 0:
			if !(strings.HasSuffix(path, ".yml") ||
				strings.HasSuffix(path, ".yaml") ||
				strings.HasSuffix(path, ".json")) {
				continue
			}
			if mode&os.ModeSymlink != 0 {
				if path, err = s.fsys.Readlink(path); err != nil {
					return err
				}
			}
			if err = s.readManifest(ctx, path); err != nil {
				return err
			}
		case mode.IsDir():
			log.Printf("Skip sub-directory %q", name)
		default:
			return fmt.Errorf("unsupported file %q with mode %q", name, mode)
		}
	}
	return nil
}

func (s *store) readManifest(ctx context.Context, path string) error {
	f, err := s.fsys.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	r := xio.NewContextReader(ctx, f)
	d := yaml.NewYAMLOrJSONDecoder(r, 4096)
	for {
		ext := runtime.RawExtension{}
		if err = d.Decode(&ext); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error parsing %s: %v", path, err)
		}

		if err = s.decode(ext.Raw); err != nil {
			return fmt.Errorf("error decode %s: %v", path, err)
		}
	}

	return nil
}

func (s *store) decode(data []byte) error {
	unstr, gvk, err := s.decoder.Decode(data, nil, nil)
	if err != nil {
		return err
	}

	var name string
	if name, err = metadataAccessor.Name(unstr); err != nil {
		return err
	} else if name == "" {
		return errors.New("missing object name")
	}

	if namespace, _ := metadataAccessor.Namespace(unstr); namespace != "" {
		log.Printf("Ignore namespace %s in object Name=%s, %s", namespace, name, gvk)
	}

	k := key{GroupVersionKind: *gvk, Name: name}
	if _, ok := s.cache[k]; ok {
		return fmt.Errorf("object %q already exists", k.Name)
	}

	obj, err := s.scheme.New(*gvk)
	if err != nil {
		return err
	}

	obj, gvk, err = s.decoder.Decode(data, gvk, obj)
	if err != nil {
		return err
	}

	s.cache[k] = obj
	return nil
}
