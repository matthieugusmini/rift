package cache

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"go.etcd.io/bbolt"

	"github.com/matthieugusmini/lolesport/rift"
)

type Cache struct {
	db  *bbolt.DB
	ttl time.Duration
}

func NewCache(db *bbolt.DB, ttl time.Duration) *Cache {
	return &Cache{
		db:  db,
		ttl: ttl,
	}
}

func (c *Cache) GetBracketTemplate(key string) (rift.BracketTemplate, bool, error) {
	f := newFacilitator[rift.BracketTemplate](c, "bracketTemplate")
	return f.get(key)
}

func (c *Cache) SetBracketTemplate(key string, value rift.BracketTemplate) error {
	f := newFacilitator[rift.BracketTemplate](c, "bracketTemplate")
	return f.set(key, value)
}

type CachedValue[T any] struct {
	Value     T     `json:"value"`
	ExpiresAt int64 `json:"expiresAt"`
}

type facilitator[T any] struct {
	cache  *Cache
	bucket string
}

func newFacilitator[T any](cache *Cache, repo string) *facilitator[T] {
	return &facilitator[T]{
		cache:  cache,
		bucket: repo,
	}
}

func (f *facilitator[T]) set(key string, value T) error {
	log.Printf("Cache - Set %s\n", key)

	expiresAt := time.Now().Add(f.cache.ttl).Unix()
	cached := CachedValue[T]{
		Value:     value,
		ExpiresAt: expiresAt,
	}

	cachedMarshaled, err := json.Marshal(cached)
	if err != nil {
		return err
	}

	err = f.cache.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(f.bucket))
		if err != nil {
			return err
		}

		err = b.Put([]byte(key), cachedMarshaled)
		return err
	})

	return err
}

func (f *facilitator[T]) get(key string) (T, bool, error) {
	var (
		zero   T
		cached CachedValue[T]
	)

	log.Printf("Cache - Get %s\n", key)

	if err := f.cache.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(f.bucket))

		if b == nil {
			return errors.New("bucket does not exist")
		}

		rawValue := b.Get([]byte(key))
		return json.Unmarshal(rawValue, &cached)
	}); err != nil {
		return zero, false, err
	}

	if cached.ExpiresAt > 0 && time.Now().Unix() > cached.ExpiresAt {
		log.Printf("Cache - Value expired %s\n", key)
		err := f.delete(key)
		return zero, false, err
	}

	return cached.Value, true, nil
}

func (f *facilitator[T]) delete(key string) error {
	log.Printf("Cache - Delete %s\n", key)

	return f.cache.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(f.bucket))
		return b.Delete([]byte(key))
	})
}
