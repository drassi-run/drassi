package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"drassi.run/core/pkg/store/cache/storage"
)

type Index interface {
	Create(ctx context.Context, cache *Cache) error
	Update(ctx context.Context, cache *Cache) error
	Get(ctx context.Context, id uint64) (*Cache, error)
	Search(ctx context.Context, keys []string, version string) (*Cache, error)
}

type manager struct {
	index   Index
	storage storage.Storage
}

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

const urlBase = "/_apis/artifactcache"

func (m *manager) Mux() *http.ServeMux {
	mux := http.NewServeMux()

	// saveCache
	mux.HandleFunc("POST "+urlBase+"/caches", m.Reserve)
	mux.HandleFunc("PATCH "+urlBase+"/caches/{id}", m.Upload)
	mux.HandleFunc("POST "+urlBase+"/caches/{id}", m.Commit)
	// restoreCache
	mux.HandleFunc("GET "+urlBase+"/cache", m.Search)
	// - custom handlers
	mux.HandleFunc("HEAD "+urlBase+"/caches/{id}", m.Metadata)
	mux.HandleFunc("GET "+urlBase+"/caches/{id}", m.Download)

	return mux
}

type reserveRequest struct {
	Key     string `json:"key"`
	Version string `json:"version"`
	Size    int64  `json:"cacheSize"`
}

// Reserve cache id
// POST /_apis/artifactcache/caches
func (m *manager) Reserve(w http.ResponseWriter, r *http.Request) {
	req := new(reserveRequest)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		m.responseJSON(w, 400, err)
		return
	} else {
		// cache keys are case-insensitive
		req.Key = strings.ToLower(req.Key)
	}

	now := time.Now()
	cache := &Cache{
		Key:        req.Key,
		Version:    req.Version,
		Size:       req.Size,
		Complete:   false,
		CreatedAt:  now,
		LastUsedAt: now,
	}
	if cache.Size == 0 {
		// So the request comes from old versions of actions, like `actions/cache@v2`.
		// It doesn't send cache size. Set it to -1 to indicate that.
		cache.Size = -1
	}

	if err := m.index.Create(r.Context(), cache); err != nil {
		m.responseJSON(w, 500, err)
	}
	m.responseJSON(w, 200, map[string]any{
		"cacheId": cache.ID,
	})
}

// Upload chunks
// PATCH /_apis/artifactcache/caches/{id}
func (m *manager) Upload(w http.ResponseWriter, r *http.Request) {
	cache := m.getCache(w, r, false)
	if cache == nil {
		return
	}

	ctx := r.Context()
	start, end, err := parseContentRangeHeader(r.Header)
	if err != nil {
		m.responseJSON(w, 400, err)
		return
	}

	length := end - start + 1
	if err = m.storage.WriteObject(ctx, cache, r.Body, start, length); err != nil {
		m.responseJSON(w, 500, err)
		return
	}

	cache.LastUsedAt = time.Now()
	_ = m.index.Update(ctx, cache)
	m.responseJSON(w, 200, nil)
}

// Commit cache upload
// POST /_apis/artifactcache/caches/{id}
func (m *manager) Commit(w http.ResponseWriter, r *http.Request) {
	cache := m.getCache(w, r, false)
	if cache == nil {
		return
	}

	ctx := r.Context()
	if err := m.storage.CommitObject(ctx, cache); err != nil {
		m.responseJSON(w, 500, err)
		return
	}

	cache.Complete = true
	cache.LastUsedAt = time.Now()
	if err := m.index.Update(ctx, cache); err != nil {
		m.responseJSON(w, 500, err)
		return
	}

	m.responseJSON(w, 200, nil)
}

type searchResponse struct {
	Key      string `json:"cacheKey"`
	Location string `json:"archiveLocation"`
	Result   string `json:"result"`
}

// Search cache by keys and version
// GET /_apis/artifactcache/cache
func (m *manager) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	version := q.Get("version")
	keys := strings.Split(q.Get("keys"), ",")
	// cache keys are case-insensitive
	for i, key := range keys {
		keys[i] = strings.ToLower(key)
	}

	ctx := r.Context()
	cache, err := m.index.Search(ctx, keys, version)
	if err != nil {
		m.responseJSON(w, 500, err)
		return
	}
	if cache == nil {
		m.responseJSON(w, 204, nil)
		return
	}

	location := m.storage.ObjectLocation(ctx, cache)
	if location == "" {
		m.responseJSON(w, 204, nil)
		return
	}

	res := &searchResponse{
		Key:      cache.Key,
		Location: location,
		Result:   "hit",
	}
	m.responseJSON(w, 200, res)
}

// Metadata return resource metadata
// HEAD /_apis/artifactcache/caches/{id}
func (m *manager) Metadata(w http.ResponseWriter, r *http.Request) {
	cache := m.getCache(w, r, true)
	if cache == nil {
		return
	}

	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(cache.Size, 10))
}

// Download cache Segment
// GET /_apis/artifactcache/caches/{id}
func (m *manager) Download(w http.ResponseWriter, r *http.Request) {
	cache := m.getCache(w, r, true)
	if cache == nil {
		return
	}

	start, end, err := parseRangeHeader(r.Header)
	if err != nil {
		m.responseJSON(w, 400, err)
		return
	}

	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(end-start+1, 10))

	length := end - start + 1
	if err = m.storage.ReadObject(r.Context(), cache, w, start, length); err != nil {
		m.responseJSON(w, 500, err)
		return
	}
}

func (m *manager) getCache(w http.ResponseWriter, r *http.Request, com bool) *Cache {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		m.responseJSON(w, 400, err)
		return nil
	}

	cache, err := m.index.Get(r.Context(), id)
	if err != nil {
		m.responseJSON(w, 500, err)
		return nil
	}

	if com != cache.Complete {
		if cache.Complete {
			err = fmt.Errorf("cache %v %q: already complete", cache.ID, cache.Key)
		} else {
			err = fmt.Errorf("cache %v %q: in-complete", cache.ID, cache.Key)
		}
		m.responseJSON(w, 400, err)
		return nil
	}

	return cache
}

var empty = struct{}{}

func (m *manager) responseJSON(w http.ResponseWriter, code int, obj any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)

	if err, ok := obj.(error); ok {
		obj = map[string]string{
			"error": err.Error(),
		}
	} else if obj == nil {
		obj = empty
	}
	_ = json.NewEncoder(w).Encode(obj)
}
