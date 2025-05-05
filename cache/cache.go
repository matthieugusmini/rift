package cache

import (
	"encoding/json"
	"errors"
	"time"

	"go.etcd.io/bbolt"
)

// Cache represents a file based cache which uses a [bbolt] file database under the hood
// for fast reads and persistency across sessions.
//
// Entries stored in the cache can be invalidated using a TTL.
//
// [bbolt]: https://github.com/etcd-io/bbolt
type Cache[T any] struct {
	db         *bbolt.DB
	bucketName string
	ttl        time.Duration
}

type entry[T any] struct {
	Value     T     `json:"value"`
	ExpiresAt int64 `json:"expiresAt"`
}

// NewCache returns a new instance of a Cache given a bbolt database, a time to live duration and a logger.
//
// If ttl == 0 values stored in the cache are never invalidated.
func NewCache[T any](db *bbolt.DB, bucketName string, ttl time.Duration) *Cache[T] {
	return &Cache[T]{
		db:         db,
		bucketName: bucketName,
		ttl:        ttl,
	}
}

// Get returns the value associated to the key in the cache.
//
// Additionally a boolean is returned to indicated whether the value was found in the cache or not.
// If the bracket template entry in the cache is subject to a TTL and has expired,
// it is invalidated and false is returned.
//
// An error is returned if the entry is corrupted or if an error happens when trying to invalidate
// the entry.
func (c *Cache[T]) Get(key string) (T, bool, error) {
	var (
		zero  T
		entry entry[T]
	)
	if err := c.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(c.bucketName))
		if bucket == nil {
			return errors.New("bucket does not exist")
		}

		b := bucket.Get([]byte(key))

		return json.Unmarshal(b, &entry)
	}); err != nil {
		return zero, false, err
	}

	hasExpired := entry.ExpiresAt > 0 && time.Now().Unix() > entry.ExpiresAt
	if hasExpired {
		err := c.delete(key)

		return zero, false, err
	}

	return entry.Value, true, nil
}

// Set stores a new entry for value in the cache associated with the given key.
//
// If the bucket used to store the rift.BracketTemplate doesn't exist in the underlying
// daabase, it is automatically created.
//
// An error is returned if cannot create a new entry or a new bucket.
func (c *Cache[T]) Set(key string, value T) error {
	expiresAt := time.Now().Add(c.ttl).Unix()
	entry := entry[T]{
		Value:     value,
		ExpiresAt: expiresAt,
	}

	b, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	if err := c.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(c.bucketName))
		if err != nil {
			return err
		}

		return bucket.Put([]byte(key), b)
	}); err != nil {
		return err
	}

	return nil
}

func (c *Cache[T]) delete(key string) error {
	return c.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(c.bucketName))
		return bucket.Delete([]byte(key))
	})
}
