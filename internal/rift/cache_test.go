package rift_test

import (
	"errors"
)

var (
	errCacheGet = errors.New(
		"DOINB RYZE HACK？英雄联盟 400 CS 24 MIN DOINB RYZE HACK？英雄联盟 400 CS 24 MIN DOINB RYZE HACK？英雄联盟 400 CS 24 MIN DOINB RYZE HACK？英雄联盟 400 CS 24 MIN DOINB RYZE HACK？英雄联盟 400 CS 24 MIN",
	)
	//nolint:staticcheck
	errCacheSet = errors.New(`It starts with success.

And success, is all about growing.

Success is all about overcoming your demons, success is about getting knocked down, time after time, and getting back up. No one embodies that more than Martin Larsson.

Rekkles. He is one of the most experienced and celebrated marksmen in Europe and the World, but regardless of that resumé, he has never won an international trophy, but he has never lost his drive.

And what's more is that even a couple of years ago, even making a Championship Final seemed out of reach.

Until 2018, everything was set up to perfection for Martin and for Fnatic. They reconquered Europe after years of G2 domination, Caps stepped up as a true carry, and they faced an IG in the Finals they had beaten twice in the group stage. But as we all know, it was a tragedy, it was a quick exit. It was a shameful display, and what was worse, it felt like a lost opportunity that may never come again, for Rekkles, for Fnatic and for Europe.

The LPL rose as the unbeateable beast we know them to be. And the LPL would dominate the World Championship two times in a row. But Rekkles, he did not stop fighting, and in fact in 2019, he was finally able to put one of his biggest demons to bed, with Uzi and RNG.

And now? Now we're here. 2020. Where the expectations for Fnatic were at their lowest ahead of this best of five versus Top Esports, the first seed from the LPL. The LPL that has beaten EU down in these best of fives, time and time again; and JackeyLove, standing tall and poised to win his second World Championship.

But so far we haven't gotten LPL domination, we haven't gotten the JackeyLove highlight reels, and we've gotten Rekkles and Hylissang lane kingdom.

It all comes down to this next game. Demons can be slayed, expectations be damned, and Rekkles can cement himself as the player he always wanted to prove he was.

In this next game. It's all up to him and Fnatic to make their dreams come true.`)
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
