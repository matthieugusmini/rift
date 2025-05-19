package cache_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/matthieugusmini/rift/internal/cache"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
)

func TestCache(t *testing.T) {
	want := "Bombardiro Crocodilo"

	t.Run("set and get success", func(t *testing.T) {
		cache := setupTestCache[string](t)

		err := cache.Set("Tralalero Tralala", want)
		require.NoError(t, err)

		got, ok, err := cache.Get("Tralalero Tralala")

		require.Equal(t, want, got)
		require.True(t, ok)
		require.NoError(t, err)
	})

	t.Run("get missing key", func(t *testing.T) {
		cache := setupTestCache[string](t)

		_, ok, err := cache.Get("Tralalero Tralala")

		require.False(t, ok)
		require.Error(t, err)
	})

	t.Run("set and get expired", func(t *testing.T) {
		cache := setupExpiringCache[string](t)

		err := cache.Set("Capuccino Assassino", "Cappucina Ballerina")
		require.NoError(t, err)

		_, ok, err := cache.Get("Capuccino Assassino")

		require.NoError(t, err)
		require.False(t, ok)
	})
}

func setupTestCache[T any](t *testing.T) *cache.Cache[T] {
	t.Helper()

	db := setupTempDB(t)
	return cache.New[T](db, "test-bucket", 0)
}

func setupExpiringCache[T any](t *testing.T) *cache.Cache[T] {
	t.Helper()

	db := setupTempDB(t)
	return cache.New[T](db, "test-bucket", -1*time.Second)
}

func setupTempDB(t *testing.T) *bbolt.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := bbolt.Open(dbPath, 0o644, nil)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	return db
}
