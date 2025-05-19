package rift_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/matthieugusmini/rift/internal/rift"
	"github.com/stretchr/testify/require"
)

func TestBracketTemplateLoader_Load(t *testing.T) {
	stageID := "42"
	want := testBracketTemplate

	t.Run("returns cached template", func(t *testing.T) {
		fakeCache := newFakeCacheWith(map[string]rift.BracketTemplate{stageID: want})
		stubAPIClient := newStubBracketTemplateAPIClient()
		loader := rift.NewBracketTemplateLoader(stubAPIClient, fakeCache, slog.Default())

		got, err := loader.Load(t.Context(), stageID)

		require.NoError(t, err)
		require.Equal(t, want, got)
	})

	t.Run("returns template from API and update cache", func(t *testing.T) {
		fakeCache := newFakeCache[rift.BracketTemplate]()
		stubAPIClient := newStubBracketTemplateAPIClient()
		loader := rift.NewBracketTemplateLoader(stubAPIClient, fakeCache, slog.Default())

		got, err := loader.Load(t.Context(), stageID)

		require.NoError(t, err)
		require.Equal(t, want, got)
		// Assert that the cache has been updated
		_, ok := fakeCache.entries[stageID]
		require.True(t, ok)
	})

	t.Run("returns error if not in cache and API not found", func(t *testing.T) {
		fakeCache := newFakeCache[rift.BracketTemplate]()
		notFoundAPIClient := newNotFoundBracketTemplateAPIClient()
		loader := rift.NewBracketTemplateLoader(notFoundAPIClient, fakeCache, slog.Default())

		_, err := loader.Load(t.Context(), stageID)

		require.Error(t, err)
	})

	t.Run("returns template even if fails to get cached value", func(t *testing.T) {
		fakeCache := newFakeCache[rift.BracketTemplate]()
		fakeCache.getErr = errCacheGet
		stubAPIClient := newStubBracketTemplateAPIClient()
		loader := rift.NewBracketTemplateLoader(stubAPIClient, fakeCache, slog.Default())

		got, err := loader.Load(t.Context(), stageID)

		require.NoError(t, err)
		require.Equal(t, want, got)
		// Assert that the cache has been updated
		_, ok := fakeCache.entries[stageID]
		require.True(t, ok)
	})

	t.Run("returns template even if cannot update cache", func(t *testing.T) {
		fakeCache := newFakeCache[rift.BracketTemplate]()
		fakeCache.setErr = errCacheSet
		stubAPIClient := newStubBracketTemplateAPIClient()
		loader := rift.NewBracketTemplateLoader(stubAPIClient, fakeCache, slog.Default())

		got, err := loader.Load(t.Context(), stageID)

		require.NoError(t, err)
		require.Equal(t, want, got)
	})
}

func TestBracketTemplateLoader_ListAvailableStageIDs(t *testing.T) {
	want := testAvailableStageIDs

	t.Run("returns stage ids", func(t *testing.T) {
		fakeCache := newFakeCache[rift.BracketTemplate]()
		stubAPIClient := newStubBracketTemplateAPIClient()
		loader := rift.NewBracketTemplateLoader(stubAPIClient, fakeCache, slog.Default())

		got, err := loader.ListAvailableStageIDs(t.Context())

		require.NoError(t, err)
		require.ElementsMatch(t, want, got)
	})

	t.Run("returns error if cannot fetch", func(t *testing.T) {
		fakeCache := newFakeCache[rift.BracketTemplate]()
		stubAPIClient := newNotFoundBracketTemplateAPIClient()
		loader := rift.NewBracketTemplateLoader(stubAPIClient, fakeCache, slog.Default())

		_, err := loader.ListAvailableStageIDs(t.Context())

		require.Error(t, err)
	})
}

var testBracketTemplate = rift.BracketTemplate{
	Rounds: []rift.Round{
		{
			Title: "test",
			Links: []rift.Link{
				{Type: rift.LinkTypeZDown, Height: 42},
			},
			Matches: []rift.Match{
				{DisplayType: rift.DisplayTypeMatch, Above: 42},
			},
		},
	},
}

var testAvailableStageIDs = []string{"1", "2"}

var errAPINotFound = errors.New(
	"체력 4700 방어력 329 마저201 인 챔피언👤이 저지불가🚫, 쉴드🛡, 벽🧱 넘기는 거 있고요. 에어본🌪 있고, 심지어 쿨타임은 1️⃣초밖에 안되고 마나🧙‍♂️는 1️⃣5️⃣ 들고 w는 심지어 변신💫하면 쿨 초기화에다가 패시브는 고정피해🗡가 들어가며 그 다음에 방마저🥋 올리면📈 올릴수록📈 스킬 가속⏰이 생기고! q에 스킬가속⏰이 생기고 스킬 속도🚀가 빨라지고📈 그 다음에 공격력🗡 계수가 있어가지고 W가 그 이익-으아아아악😱😱---",
)

type stubBracketTemplateAPIClient struct {
	template          rift.BracketTemplate
	availableStageIDs []string
	err               error
}

func newStubBracketTemplateAPIClient() *stubBracketTemplateAPIClient {
	return &stubBracketTemplateAPIClient{
		template:          testBracketTemplate,
		availableStageIDs: testAvailableStageIDs,
	}
}

func newNotFoundBracketTemplateAPIClient() *stubBracketTemplateAPIClient {
	return &stubBracketTemplateAPIClient{err: errAPINotFound}
}

func (api *stubBracketTemplateAPIClient) GetTemplateByStageID(
	_ context.Context,
	_ string,
) (rift.BracketTemplate, error) {
	if api.err != nil {
		return rift.BracketTemplate{}, api.err
	}
	return api.template, nil
}

func (api *stubBracketTemplateAPIClient) ListAvailableStageIDs(
	_ context.Context,
) ([]string, error) {
	if api.err != nil {
		return nil, api.err
	}
	return api.availableStageIDs, nil
}
