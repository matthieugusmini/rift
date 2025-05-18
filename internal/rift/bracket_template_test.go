package rift_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/matthieugusmini/rift/internal/rift"
)

func TestBracketTemplateLoader_Load(t *testing.T) {
	stageID := "Chunin Exams"
	want := testBracketTemplate

	t.Run("when template present in cache returns bracket template", func(t *testing.T) {
		fakeCache := newFakeCacheWith(map[string]rift.BracketTemplate{stageID: want})
		stubAPIClient := newStubBracketTemplateAPIClient()
		loader := rift.NewBracketTemplateLoader(stubAPIClient, fakeCache, slog.Default())

		got := mustLoadBracketTemplate(t, loader, stageID)

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf(
				"BracketTemplateLoader.Load(stageID) returned unexpected diffs(-want +got):\n%s",
				diff,
			)
		}
	})

	t.Run(
		"when template not in cache fetch from API and returns bracket template and set value in cache",
		func(t *testing.T) {
			fakeCache := newFakeCache[rift.BracketTemplate]()
			stubAPIClient := newStubBracketTemplateAPIClient()
			loader := rift.NewBracketTemplateLoader(stubAPIClient, fakeCache, slog.Default())

			got := mustLoadBracketTemplate(t, loader, stageID)

			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf(
					"BracketTemplateLoader.Load(stageID) returned unexpected diffs(-want +got):\n%s",
					diff,
				)
			}

			// Assert that the cache has been updated
			cacheEntry, ok := fakeCache.entries[stageID]
			if !ok {
				t.Fatalf("Bracket template should be cached after loading")
			}
			if diff := cmp.Diff(want, cacheEntry); diff != "" {
				t.Errorf("Cache[stageID] has unexpected diffs(-want +got):\n%s", diff)
			}
		},
	)

	t.Run("when template not in cache and not found in API returns error", func(t *testing.T) {
		fakeCache := newFakeCache[rift.BracketTemplate]()
		notFoundAPIClient := newNotFoundBracketTemplateAPIClient()
		loader := rift.NewBracketTemplateLoader(notFoundAPIClient, fakeCache, slog.Default())

		_, err := loader.Load(t.Context(), stageID)
		if err == nil {
			t.Error("expected an error, got nil")
		}
	})

	t.Run(
		"when fails to get in cache should log and returns bracket template",
		func(t *testing.T) {
			fakeCache := &fakeCache[rift.BracketTemplate]{
				entries: map[string]rift.BracketTemplate{},
				getErr:  errCacheGet,
			}
			stubAPIClient := newStubBracketTemplateAPIClient()
			loader := rift.NewBracketTemplateLoader(stubAPIClient, fakeCache, slog.Default())

			got := mustLoadBracketTemplate(t, loader, stageID)

			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf(
					"BracketTemplateLoader.Load(stageID) returned unexpected diffs(-want +got):\n%s",
					diff,
				)
			}

			// Assert that the cache has been updated
			cacheEntry, ok := fakeCache.entries[stageID]
			if !ok {
				t.Fatalf("Bracket template should be cached after loading")
			}
			if diff := cmp.Diff(want, cacheEntry); diff != "" {
				t.Errorf("Cache[stageID] has unexpected diffs(-want +got):\n%s", diff)
			}
		},
	)

	t.Run("when fails to update cache should log and return bracket template", func(t *testing.T) {
		fakeCache := &fakeCache[rift.BracketTemplate]{
			entries: map[string]rift.BracketTemplate{},
			setErr:  errCacheSet,
		}
		stubAPIClient := newStubBracketTemplateAPIClient()
		loader := rift.NewBracketTemplateLoader(stubAPIClient, fakeCache, slog.Default())

		got := mustLoadBracketTemplate(t, loader, stageID)

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf(
				"BracketTemplateLoader.Load(stageID) returned unexpected diffs(-want +got):\n%s",
				diff,
			)
		}
	})
}

func TestBracketTemplateLoader_ListAvailableStageIDs(t *testing.T) {
	want := testAvailableStageIDs

	t.Run("returns available stage ids when successfully fetch", func(t *testing.T) {
		fakeCache := newFakeCache[rift.BracketTemplate]()
		stubAPIClient := newStubBracketTemplateAPIClient()
		loader := rift.NewBracketTemplateLoader(stubAPIClient, fakeCache, slog.Default())

		got, err := loader.ListAvailableStageIDs(t.Context())
		if err != nil {
			t.Fatalf("got unexpected error %q, want nil", err)
		}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf(
				"BracketTemplateLoader.ListAvailableStageIDs() returned unexpected diffs(-want +got):\n%s",
				diff,
			)
		}
	})

	t.Run("returns error when cannot fetch", func(t *testing.T) {
		fakeCache := newFakeCache[rift.BracketTemplate]()
		stubAPIClient := newNotFoundBracketTemplateAPIClient()
		loader := rift.NewBracketTemplateLoader(stubAPIClient, fakeCache, slog.Default())

		_, err := loader.ListAvailableStageIDs(t.Context())
		if err == nil {
			t.Error("expected an error, got nil")
		}
	})
}

func mustLoadBracketTemplate(
	t *testing.T,
	loader *rift.BracketTemplateLoader,
	stageID string,
) rift.BracketTemplate {
	t.Helper()

	got, err := loader.Load(t.Context(), stageID)
	if err != nil {
		t.Fatalf("got unexpected error %q, want nil", err)
	}

	return got
}

var testBracketTemplate = rift.BracketTemplate{
	Rounds: []rift.Round{
		{
			Title: "Quarterfinals Worlds 2020",
			Links: []rift.Link{
				{Type: rift.LinkTypeZDown, Height: 42},
			},
			Matches: []rift.Match{
				{DisplayType: rift.DisplayTypeMatch, Above: 42},
			},
		},
	},
}

var testAvailableStageIDs = []string{"caliste", "vladi"}

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
