package githubusercontent_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matthieugusmini/rift/internal/githubusercontent"
	"github.com/matthieugusmini/rift/internal/rift"
	"github.com/stretchr/testify/require"
)

func TestBracketTemplateClient_ListAvailableStageIDs(t *testing.T) {
	response := testBracketTypeByStageID
	want := []string{"1", "2"}

	t.Run("returns available stage ids when successful request", func(t *testing.T) {
		client, mux := setup(t)
		mux.HandleFunc(
			"/bracket-type-by-stage-id.json",
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				json.NewEncoder(w).Encode(response) //nolint:errcheck
			},
		)

		got, err := client.ListAvailableStageIDs(t.Context())

		require.NoError(t, err)
		require.ElementsMatch(t, want, got)
	})

	t.Run("returns error if malformated JSON", func(t *testing.T) {
		client, mux := setup(t)
		mux.HandleFunc(
			"/bracket-type-by-stage-id.json",
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				fmt.Fprint(w, "malformated JSON")
			},
		)

		_, err := client.ListAvailableStageIDs(t.Context())

		require.Error(t, err)
	})
}

func TestBracketTemplateClient_GetTemplateByStageID(t *testing.T) {
	stageID := "1"
	response := testBracketTypeByStageID
	want := testBracketTemplate

	t.Run("returns bracket template when successful request", func(t *testing.T) {
		client, mux := setup(t)
		mux.HandleFunc(
			"/bracket-type-by-stage-id.json",
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				json.NewEncoder(w).Encode(response) //nolint:errcheck
			}),
		)
		mux.HandleFunc("/8SE.json", func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			json.NewEncoder(w).Encode(want) //nolint:errcheck
		})

		got, err := client.GetTemplateByStageID(t.Context(), stageID)

		require.NoError(t, err)
		require.Equal(t, want, got)
	})

	t.Run("returns error when stage id is not supported", func(t *testing.T) {
		client, mux := setup(t)
		mux.HandleFunc(
			"/bracket-type-by-stage-id.json",
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				json.NewEncoder(w).Encode(response) //nolint:errcheck
			}),
		)
		_, err := client.GetTemplateByStageID(t.Context(), "3")

		require.Error(t, err)
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
