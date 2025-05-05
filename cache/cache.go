package cache

import (
	"encoding/json"
	"errors"
	"time"

	"go.etcd.io/bbolt"

	"github.com/matthieugusmini/lolesport/rift"
)

// Cache represents a file based cache which uses a [bbolt] file database under the hood
// for fast reads and persistency across sessions.
//
// Entries stored in the cache can be invalidated using a TTL.
//
// [bbolt]: https://github.com/etcd-io/bbolt
type Cache struct {
	db  *bbolt.DB
	ttl time.Duration
}

// NewCache returns a new instance of a Cache given a bbolt database, a time to live duration and a logger.
//
// If ttl == 0 values stored in the cache are never invalidated.
func NewCache(db *bbolt.DB, ttl time.Duration) *Cache {
	return &Cache{
		db:  db,
		ttl: ttl,
	}
}

// GetBracketTemplate returns the rift.BracketTemplate associated to the key in the cache.
//
// Additionally a boolean is returned to indicated whether the value was found in the cache or not.
// If the bracket template entry in the cache is subject to a TTL and has expired,
// it is invalidated and false is returned.
//
// An error is returned if the entry is corrupted or if an error happens when trying to invalidate
// the entry.
func (c *Cache) GetBracketTemplate(key string) (rift.BracketTemplate, bool, error) {
	f := newFacilitator[rift.BracketTemplate](c, "bracketTemplate")
	return f.get(key)
}

// SetBracketTemplate store a new rift.BracketTemplate entry in the cache associated with the given key.
//
// If the bucket used to store the rift.BracketTemplate doesn't exist in the underlying
// daabase, it is automatically created.
//
// An error is returned if cannot create a new entry or a new bucket.
func (c *Cache) SetBracketTemplate(key string, value rift.BracketTemplate) error {
	f := newFacilitator[rift.BracketTemplate](c, "bracketTemplate")
	return f.set(key, value)
}

type entry[T any] struct {
	Value     T     `json:"value"`
	ExpiresAt int64 `json:"expiresAt"`
}

type facilitator[T any] struct {
	cache  *Cache
	bucket string
}

func newFacilitator[T any](cache *Cache, bucket string) *facilitator[T] {
	return &facilitator[T]{
		cache:  cache,
		bucket: bucket,
	}
}

func (f *facilitator[T]) set(key string, value T) error {
	expiresAt := time.Now().Add(f.cache.ttl).Unix()
	entry := entry[T]{
		Value:     value,
		ExpiresAt: expiresAt,
	}

	b, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	if err := f.cache.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(f.bucket))
		if err != nil {
			return err
		}

		return bucket.Put([]byte(key), b)
	}); err != nil {
		return err
	}

	return nil
}

func (f *facilitator[T]) get(key string) (T, bool, error) {
	var (
		zero   T
		cached entry[T]
	)
	if err := f.cache.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(f.bucket))
		if bucket == nil {
			return errors.New("bucket does not exist")
		}

		b := bucket.Get([]byte(key))

		return json.Unmarshal(b, &cached)
	}); err != nil {
		return zero, false, err
	}

	hasExpired := cached.ExpiresAt > 0 && time.Now().Unix() > cached.ExpiresAt
	if hasExpired {
		err := f.delete(key)

		return zero, false, err
	}

	return cached.Value, true, nil
}

func (f *facilitator[T]) delete(key string) error {
	return f.cache.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(f.bucket))
		return bucket.Delete([]byte(key))
	})
}
