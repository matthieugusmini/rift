package githubusercontent_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matthieugusmini/rift/internal/githubusercontent"
	"github.com/matthieugusmini/rift/internal/rift"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBracketTemplateClient_ListAvailableStageIDs(t *testing.T) {
	response := testBracketTypeByStageID
	want := []string{"1", "2"}

	t.Run("successful request returns available stage ids", func(t *testing.T) {
		client, mux := setup(t)
		mux.HandleFunc(
			"/bracket-type-by-stage-id.json",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)

				json.NewEncoder(w).Encode(response)
			},
		)

		got, err := client.ListAvailableStageIDs(t.Context())

		require.NoError(t, err)
		assert.ElementsMatch(t, want, got)
	})

	t.Run("malformated JSON return error", func(t *testing.T) {
		client, mux := setup(t)
		mux.HandleFunc(
			"/bracket-type-by-stage-id.json",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)

				fmt.Fprint(w, "malformated JSON")
			},
		)

		_, err := client.ListAvailableStageIDs(t.Context())

		assert.Error(t, err)
	})

	t.Run("status not OK returns error", func(t *testing.T) {
		client, mux := setup(t)
		mux.HandleFunc(
			"/bracket-type-by-stage-id.json",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)

				http.NotFound(w, r)
			},
		)

		_, err := client.ListAvailableStageIDs(t.Context())

		assert.Error(t, err)
	})
}

func TestBracketTemplateClient_GetTemplateByStageID(t *testing.T) {
	stageID := "1"
	response := testBracketTypeByStageID
	want := testBracketTemplate

	t.Run("successful request returns bracket template", func(t *testing.T) {
		client, mux := setup(t)

		mux.HandleFunc(
			"/bracket-type-by-stage-id.json",
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)

				json.NewEncoder(w).Encode(response)
			}),
		)
		mux.HandleFunc("/8SE.json", func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)

			json.NewEncoder(w).Encode(want)
		})

		got, err := client.GetTemplateByStageID(t.Context(), stageID)

		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("stage id not supported returns error", func(t *testing.T) {
		client, mux := setup(t)
		mux.HandleFunc(
			"/bracket-type-by-stage-id.json",
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)

				json.NewEncoder(w).Encode(response)
			}),
		)
		_, err := client.GetTemplateByStageID(t.Context(), "3")

		assert.Error(t, err)
	})

	t.Run("status not OK returns error", func(t *testing.T) {
		client, mux := setup(t)
		mux.HandleFunc(
			"/bracket-type-by-stage-id.json",
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)

				http.NotFound(w, r)
			}),
		)
		_, err := client.GetTemplateByStageID(t.Context(), "3")

		assert.Error(t, err)
	})
}

func setup(
	t *testing.T,
) (*githubusercontent.BracketTemplateClient, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := githubusercontent.NewBracketTemplateClient(
		http.DefaultClient,
		githubusercontent.WithBaseURL(srv.URL),
	)

	return client, mux
}

var testBracketTypeByStageID = map[string]string{
	"1": "8SE",
	"2": "4SE",
}

var testBracketTemplate = rift.BracketTemplate{}
