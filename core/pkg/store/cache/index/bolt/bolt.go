package bolt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"drassi.run/core/pkg/store/cache/index"
	"drassi.run/core/pkg/store/cache/types"
	"drassi.run/core/util/fs"
	"drassi.run/core/util/path"
	"github.com/timshannon/bolthold"
	"go.etcd.io/bbolt"
)

type cache struct {
	ID         uint64    `json:"id" boltholdKey:"Id"`
	Key        string    `json:"key" boltholdKey:"Key"`
	Version    string    `json:"version" boltholdKey:"Version"`
	Namespace  string    `json:"namespace" boltholdKey:"Namespace"`
	Size       int64     `json:"cacheSize" boltholdKey:"CacheSize"`
	Complete   bool      `json:"complete" boltholdKey:"Complete"`
	CreatedAt  time.Time `json:"createdAt" boltholdKey:"CreatedAt"`
	LastUsedAt time.Time `json:"lastUsedAt" boltholdKey:"LastUsedAt"`
}

type idxBolt struct {
	db *bolthold.Store
}

func New(dir string) (index.Index, error) {
	if d, err := xpath.ResolveDir(dir); err != nil {
		return nil, err
	} else {
		dir = d
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	file := filepath.Join(dir, "bolt.db")
	opts := &bolthold.Options{
		Encoder: json.Marshal,
		Decoder: json.Unmarshal,
		Options: &bbolt.Options{
			Timeout:      5 * time.Second,
			NoGrowSync:   bbolt.DefaultOptions.NoGrowSync,
			FreelistType: bbolt.DefaultOptions.FreelistType,
		},
	}
	if db, err := bolthold.Open(file, xfs.FilePerm, opts); err != nil {
		return nil, err
	} else {
		idx := &idxBolt{db: db}
		return idx, nil
	}
}

func (i *idxBolt) Create(_ context.Context, c *types.Cache) error {
	m := cache(*c)

	if err := i.db.Insert(bolthold.NextSequence(), &m); err != nil {
		return fmt.Errorf("insert cache: %w", err)
	}
	// write back id to db
	if err := i.db.Update(m.ID, &m); err != nil {
		return fmt.Errorf("write back id to db: %w", err)
	}

	c.ID = m.ID
	return nil
}

func (i *idxBolt) Update(_ context.Context, c *types.Cache) error {
	m := cache(*c)
	return i.db.Update(m.ID, &m)
}

func (i *idxBolt) Get(_ context.Context, id uint64) (*types.Cache, error) {
	c := new(cache)
	if err := i.db.Get(id, c); err != nil {
		return nil, err
	} else {
		m := types.Cache(*c)
		return &m, nil
	}
}

func (i *idxBolt) Search(_ context.Context, keys []string, version string) (*types.Cache, error) {
	c := new(cache)

	for query := range i.searchQueries(keys, version) {
		if err := i.db.FindOne(c, query); err != nil {
			if !errors.Is(err, bolthold.ErrNotFound) {
				return nil, err
			}
		} else {
			m := types.Cache(*c)
			return &m, nil
		}
	}

	return nil, nil
}

func (i *idxBolt) searchQueries(keys []string, version string) iter.Seq[*bolthold.Query] {
	baseQuery := func() *bolthold.Query {
		return bolthold.
			Where("Version").Eq(version).
			And("Complete").Eq(true).
			SortBy("CreatedAt").Reverse()
	}

	return func(yield func(*bolthold.Query) bool) {
		for _, key := range keys {
			query := baseQuery().And("Key").Eq(key)
			if !yield(query) {
				return
			}

			if re, err := regexp.Compile("^" + regexp.QuoteMeta(key)); err != nil {
				continue
			} else {
				query = baseQuery().And("Key").RegExp(re)
			}
			if !yield(query) {
				return
			}
		}
	}
}
