package githubusercontent_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/matthieugusmini/rift/internal/githubusercontent"
	"github.com/matthieugusmini/rift/internal/rift"
)

func setupBracketTemplateClientAndServer(
	t *testing.T,
) (*githubusercontent.BracketTemplateClient, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()

	bracketTypeByStageID := map[string]string{
		"1": "8SE",
		"2": "4SE",
	}
	mux.HandleFunc(
		"/bracket-type-by-stage-id.json",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := json.NewEncoder(w).Encode(bracketTypeByStageID); err != nil {
				t.Fatalf("failed to encode JSON: %v", err)
			}
		}),
	)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := githubusercontent.NewBracketTemplateClient(
		http.DefaultClient,
		githubusercontent.WithBaseURL(srv.URL),
	)

	return client, mux
}

func TestBracketTemplateClient_ListAvailableStageIDs(t *testing.T) {
	want := []string{"1", "2"}

	t.Run("returns available stage ids when successful request", func(t *testing.T) {
		client, _ := setupBracketTemplateClientAndServer(t)

		got, err := client.ListAvailableStageIDs(t.Context())
		if err != nil {
			t.Fatalf("got unexpected error %q, want nil", err)
		}

		if !containsAll(got, want) {
			t.Errorf("BracketTemplateClient.ListAvailableStageIDs() = %v, want %v", got, want)
		}
	})
}

func TestBracketTemplateClient_GetTemplateByStageID(t *testing.T) {
	stageID := "1"
	want := testBracketTemplate

	t.Run("returns bracket template when successful request", func(t *testing.T) {
		client, mux := setupBracketTemplateClientAndServer(t)

		mux.HandleFunc("/8SE.json", func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(want)
		})

		got, err := client.GetTemplateByStageID(t.Context(), stageID)
		if err != nil {
			t.Fatalf("got unexpected error %q, want nil", err)
		}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf(
				"BracketTemplateClient.GetTemplateByStageID(stageID) returned unexpected diffs(-want +got):\n%s",
				diff,
			)
		}
	})

	t.Run("returns error when stage id is not supported", func(t *testing.T) {
		client, _ := setupBracketTemplateClientAndServer(t)

		_, err := client.GetTemplateByStageID(t.Context(), "3")
		if err == nil {
			t.Error("expected an error, got nil")
		}
	})
}

func containsAll[T comparable](haystack, needles []T) bool {
	seen := map[T]bool{}

	for _, elem := range haystack {
		seen[elem] = true
	}

	for _, v := range needles {
		if _, ok := seen[v]; !ok {
			return false
		}
	}

	return true
}

var testBracketTemplate = rift.BracketTemplate{}
