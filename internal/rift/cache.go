package rift

// Cache represents a key/value store with fast read access.
type Cache[T any] interface {
	// Get should return:
	// - The value associated with the key that is retrieved from the cache
	// - A boolean indicating whether the value was found or not in the cache
	// - An error(e.g. could not invalidate entry, etc.)
	Get(key string) (T, bool, error)

	// Set should create a new entry with key and value in the cache.
	Set(key string, value T) error
}
