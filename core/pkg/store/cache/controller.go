package cache

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"drassi.run/core/pkg/store/cache/index"
	"drassi.run/core/pkg/store/cache/storage"
	"drassi.run/core/pkg/store/cache/types"
)

type Controller interface {
	Mux() *http.ServeMux
}

func New(idx index.Index, s storage.Storage) Controller {
	return &controller{
		index:   idx,
		storage: s,
	}
}

type controller struct {
	index   index.Index
	storage storage.Storage
}

const urlBase = "/_apis/artifactcache"

func (c *controller) Mux() *http.ServeMux {
	mux := http.NewServeMux()

	// saveCache
	mux.HandleFunc("POST "+urlBase+"/caches", c.Reserve)
	mux.HandleFunc("PATCH "+urlBase+"/caches/{id}", c.Upload)
	mux.HandleFunc("POST "+urlBase+"/caches/{id}", c.Commit)
	// restoreCache
	mux.HandleFunc("GET "+urlBase+"/cache", c.Search)
	// - custom handlers
	mux.HandleFunc("HEAD "+urlBase+"/caches/{id}", c.Metadata)
	mux.HandleFunc("GET "+urlBase+"/caches/{id}", c.Download)

	return mux
}

type reserveRequest struct {
	Key     string `json:"key"`
	Version string `json:"version"`
	Size    int64  `json:"cacheSize"`
}

// Reserve cache id
// POST /_apis/artifactcache/caches
func (c *controller) Reserve(w http.ResponseWriter, r *http.Request) {
	req := new(reserveRequest)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		c.responseJSON(w, 400, err)
		return
	} else {
		// cache keys are case-insensitive
		req.Key = strings.ToLower(req.Key)
	}

	now := time.Now()
	cache := &types.Cache{
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

	ctx := r.Context()
	if err := c.index.Create(ctx, cache); err != nil {
		c.responseJSON(w, 500, err)
	}
	if err := c.storage.InitObject(ctx, cache); err != nil {
		c.responseJSON(w, 500, err)
	}

	c.responseJSON(w, 200, map[string]any{
		"cacheId": cache.ID,
	})
}

// Upload chunks
// PATCH /_apis/artifactcache/caches/{id}
func (c *controller) Upload(w http.ResponseWriter, r *http.Request) {
	cache := c.getCache(w, r, false)
	if cache == nil {
		return
	}

	ctx := r.Context()
	start, end, err := parseContentRangeHeader(r.Header)
	if err != nil {
		c.responseJSON(w, 400, err)
		return
	}

	length := end - start + 1
	if err = c.storage.WriteObject(ctx, cache, r.Body, start, length); err != nil {
		c.responseJSON(w, 500, err)
		return
	}

	cache.LastUsedAt = time.Now()
	_ = c.index.Update(ctx, cache)
	c.responseJSON(w, 200, nil)
}

// Commit cache upload
// POST /_apis/artifactcache/caches/{id}
func (c *controller) Commit(w http.ResponseWriter, r *http.Request) {
	cache := c.getCache(w, r, false)
	if cache == nil {
		return
	}

	ctx := r.Context()
	if err := c.storage.CommitObject(ctx, cache); err != nil {
		c.responseJSON(w, 500, err)
		return
	}

	cache.Complete = true
	cache.LastUsedAt = time.Now()
	if err := c.index.Update(ctx, cache); err != nil {
		c.responseJSON(w, 500, err)
		return
	}

	c.responseJSON(w, 200, nil)
}

type searchResponse struct {
	Key      string `json:"cacheKey"`
	Location string `json:"archiveLocation"`
	Result   string `json:"result"`
}

// Search cache by keys and version
// GET /_apis/artifactcache/cache
func (c *controller) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	version := q.Get("version")
	keys := strings.Split(q.Get("keys"), ",")
	// cache keys are case-insensitive
	for i, key := range keys {
		keys[i] = strings.ToLower(key)
	}

	ctx := r.Context()
	cache, err := c.index.Search(ctx, keys, version)
	if err != nil {
		c.responseJSON(w, 500, err)
		return
	}
	if cache == nil {
		c.responseJSON(w, 204, nil)
		return
	}

	location := c.storage.ObjectLocation(ctx, cache)
	if location == "" {
		location = fmt.Sprintf("%s/_apis/artifactcache/caches/%d", c.requestOrigin(r), cache.ID)
	}

	res := &searchResponse{
		Key:      cache.Key,
		Location: location,
		Result:   "hit",
	}
	c.responseJSON(w, 200, res)
}

// Metadata return resource metadata
// HEAD /_apis/artifactcache/caches/{id}
func (c *controller) Metadata(w http.ResponseWriter, r *http.Request) {
	cache := c.getCache(w, r, true)
	if cache == nil {
		return
	}

	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(cache.Size, 10))
}

// Download cache Segment
// GET /_apis/artifactcache/caches/{id}
func (c *controller) Download(w http.ResponseWriter, r *http.Request) {
	cache := c.getCache(w, r, true)
	if cache == nil {
		return
	}

	start, end, err := parseRangeHeader(r.Header)
	if err != nil {
		c.responseJSON(w, 400, err)
		return
	}

	length := cache.Size - start
	if 0 < end {
		length = end - start + 1
	}

	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(length, 10))

	ctx := r.Context()
	cache.LastUsedAt = time.Now()
	if err = c.index.Update(ctx, cache); err != nil {
		c.responseJSON(w, 500, err)
		return
	}
	if err = c.storage.ReadObject(ctx, cache, w, start, length); err != nil {
		c.responseJSON(w, 500, err)
		return
	}
}

func (c *controller) getCache(w http.ResponseWriter, r *http.Request, com bool) *types.Cache {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		c.responseJSON(w, 400, err)
		return nil
	}

	cache, err := c.index.Get(r.Context(), id)
	if err != nil {
		c.responseJSON(w, 500, err)
		return nil
	}

	if com != cache.Complete {
		if cache.Complete {
			err = fmt.Errorf("cache %v %q: already complete", cache.ID, cache.Key)
		} else {
			err = fmt.Errorf("cache %v %q: in-complete", cache.ID, cache.Key)
		}
		c.responseJSON(w, 400, err)
		return nil
	}

	return cache
}

var empty = struct{}{}

func (c *controller) responseJSON(w http.ResponseWriter, code int, obj any) {
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

func (c *controller) requestOrigin(r *http.Request) string {
	o := "http"
	if r.TLS != nil {
		o = "https"
	}
	return o + "://" + r.Host
}
