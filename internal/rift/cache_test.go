package rift_test

import (
	"errors"
)

var (
	errCacheGet = errors.New("could not get value in cache")
	errCacheSet = errors.New("could not update value in cache")
)

type fakeCache[T any] struct {
	entries map[string]T
	getErr  error
	setErr  error
}

func newFakeCache[T any]() *fakeCache[T] {
	return &fakeCache[T]{
		entries: map[string]T{},
	}
}

func newFakeCacheWith[T any](entries map[string]T) *fakeCache[T] {
	return &fakeCache[T]{entries: entries}
}

func (c *fakeCache[T]) Get(key string) (T, bool, error) {
	if c.getErr != nil {
		return *(new(T)), false, c.getErr
	}

	tmpl, ok := c.entries[key]
	if !ok {
		return *(new(T)), false, nil
	}
	return tmpl, true, nil
}

func (c *fakeCache[T]) Set(key string, value T) error {
	if c.setErr != nil {
		return c.setErr
	}

	c.entries[key] = value
	return nil
}
