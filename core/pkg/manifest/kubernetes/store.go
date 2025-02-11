package kubernetes

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"drassi.run/core/pkg/manifest"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	diskcached "k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/util/homedir"
)

type store struct {
	client     rest.Interface // See [k8s.io/client-go/gentype.Client]
	mapper     meta.RESTMapper
	scheme     *runtime.Scheme
	namespace  string
	paramCodec runtime.ParameterCodec
}

func NewStore(client rest.Interface, mapper meta.RESTMapper, scheme *runtime.Scheme, namespace string) manifest.Store {
	s := &store{
		client:     client,
		mapper:     mapper,
		scheme:     scheme,
		namespace:  namespace,
		paramCodec: runtime.NewParameterCodec(scheme),
	}
	return s
}

func NewStoreForConfig(config *rest.Config, scheme *runtime.Scheme, namespace string) (manifest.Store, error) {
	client, err := rest.RESTClientFor(config)
	if err != nil {
		return nil, err
	}

	// see [k8s.io/cli-runtime/pkg/genericclioptions.ConfigFlags#ToDiscoveryClient]
	cacheDir := getDefaultCacheDir()
	httpCacheDir := filepath.Join(cacheDir, "http")
	discoveryCacheDir := computeDiscoverCacheDir(filepath.Join(cacheDir, "discovery"), config.Host)
	discoveryClient, err := diskcached.NewCachedDiscoveryClientForConfig(config, discoveryCacheDir, httpCacheDir, 6*time.Hour)
	if err != nil {
		return nil, err
	}

	// see [k8s.io/cli-runtime/pkg/genericclioptions.ConfigFlags#ToRESTMapper]
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient)

	s := NewStore(client, mapper, scheme, namespace)
	return s, nil
}

func getDefaultCacheDir() string {
	if kcd := os.Getenv("KUBECACHEDIR"); kcd != "" {
		return kcd
	}

	return filepath.Join(homedir.HomeDir(), ".kube", "cache")
}

// overlyCautiousIllegalFileCharacters matches characters that *might* not be supported.  Windows is really restrictive, so this is really restrictive
var overlyCautiousIllegalFileCharacters = regexp.MustCompile(`[^(\w/.)]`)
var schemeHttp = regexp.MustCompile("^https?://")

// computeDiscoverCacheDir takes the parentDir and the host and comes up with a "usually non-colliding" name.
func computeDiscoverCacheDir(parentDir, host string) string {
	// strip the optional scheme from host if it's there:
	schemelessHost := schemeHttp.ReplaceAllString(host, "")
	// now do a simple collapse of non-AZ09 characters.  Collisions are possible but unlikely.  Even if we do collide the problem is short-lived
	safeHost := overlyCautiousIllegalFileCharacters.ReplaceAllString(schemelessHost, "_")
	return filepath.Join(parentDir, safeHost)
}

// Load takes name and gvk of the resource, and returns the corresponding object
// see [k8s.io/client-go/gentype.Client#Get]
func (s *store) Load(ctx context.Context, gvk schema.GroupVersionKind, name string) (runtime.Object, error) {
	o, err := s.scheme.New(gvk)
	if err != nil {
		return nil, err
	}
	mapping, err := s.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}

	err = s.client.Get().
		UseProtobufAsDefaultIfPreferred(true).
		NamespaceIfScoped(s.namespace, s.namespace != "").
		Resource(mapping.Resource.Resource).
		Name(name).
		VersionedParams(new(metav1.GetOptions), s.paramCodec).
		Do(ctx).
		Into(o)
	if err != nil {
		return nil, err
	}

	s.scheme.Default(o)
	return o, nil
}
