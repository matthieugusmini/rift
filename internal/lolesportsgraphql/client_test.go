package lolesportsgraphql_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/matthieugusmini/rift/internal/lolesportsgraphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_GetStage(t *testing.T) {
	t.Run("returns a stage with its bracket topology", func(t *testing.T) {
		client, mux := setup(t)
		mux.HandleFunc("/api/gql", func(w http.ResponseWriter, r *http.Request) {
			assertGetStageRequest(t, r, "stage-1")

			fixture, err := os.ReadFile("testdata/get_season_stage.json")
			if !assert.NoError(t, err) {
				http.Error(w, "could not read fixture", http.StatusInternalServerError)

				return
			}

			_, err = w.Write(fixture)
			assert.NoError(t, err)
		})

		stage, err := client.GetStage(t.Context(), "stage-1")

		require.NoError(t, err)
		assert.Equal(t, "stage-1", stage.ID)
		assert.Equal(t, "Playoffs", stage.Name)
		assert.Equal(t, lolesportsgraphql.DestinationTypeStanding, stage.Destinations[0].Type)
		require.Len(t, stage.Sections, 1)
		assert.Equal(t, lolesportsgraphql.SectionTypeBracket, stage.Sections[0].Type)
		require.Len(t, stage.Sections[0].Columns, 2)

		semifinal := stage.Sections[0].Columns[0].Cells[0].Matches[0]
		assert.Equal(t, "semifinal-1", semifinal.StructuralID)
		require.NotNil(t, semifinal.Destinations)
		assert.Equal(t, lolesportsgraphql.DestinationTypeMatch, semifinal.Destinations.Win.Type)
		assert.Equal(t, new("final"), semifinal.Destinations.Win.StructuralID)
		require.NotNil(t, semifinal.Teams[0].Result)
		assert.Equal(t, new("win"), semifinal.Teams[0].Result.Outcome)

		final := stage.Sections[0].Columns[1].Cells[0].Matches[0]
		assert.Nil(t, final.Destinations)
		assert.Empty(t, final.Teams)
		require.NotNil(t, stage.Sections[0].Standings[0].Team.Record)
		assert.Equal(t, 2, stage.Sections[0].Standings[0].Team.Record.Wins)
	})

	t.Run("returns GraphQL errors", func(t *testing.T) {
		client, mux := setup(t)
		mux.HandleFunc("/api/gql", func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, `{"errors":[{"message":"persisted query expired"}]}`)
		})

		_, err := client.GetStage(t.Context(), "stage-1")

		require.ErrorContains(t, err, "persisted query expired")
	})

	t.Run("identifies an unavailable persisted query", func(t *testing.T) {
		client, mux := setup(t)
		mux.HandleFunc("/api/gql", func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, `{
                "errors":[{
                    "message":"PersistedQueryNotInList",
                    "extensions":{"code":"PERSISTED_QUERY_NOT_IN_LIST"}
                }]
            }`)
		})

		_, err := client.GetStage(t.Context(), "stage-1")

		require.ErrorIs(t, err, lolesportsgraphql.ErrPersistedQueryUnavailable)
	})

	t.Run("returns an error when the stage is absent", func(t *testing.T) {
		client, mux := setup(t)
		mux.HandleFunc("/api/gql", func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, `{"data":{"stages":[]}}`)
		})

		_, err := client.GetStage(t.Context(), "missing")

		require.ErrorContains(t, err, "stage not found")
	})

	t.Run("returns an error for a non-OK response", func(t *testing.T) {
		client, mux := setup(t)
		mux.HandleFunc("/api/gql", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "unavailable", http.StatusServiceUnavailable)
		})

		_, err := client.GetStage(t.Context(), "stage-1")

		require.ErrorContains(t, err, "unexpected status code: 503")
	})

	t.Run("returns an error for malformed JSON", func(t *testing.T) {
		client, mux := setup(t)
		mux.HandleFunc("/api/gql", func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, "malformed JSON")
		})

		_, err := client.GetStage(t.Context(), "stage-1")

		require.ErrorContains(t, err, "could not decode response body")
	})
}

func setup(t *testing.T) (*lolesportsgraphql.Client, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := lolesportsgraphql.NewClient(
		server.Client(),
		lolesportsgraphql.WithBaseURL(server.URL+"/api/gql"),
	)

	return client, mux
}

func assertGetStageRequest(t *testing.T, r *http.Request, stageID string) {
	t.Helper()

	assert.Equal(t, http.MethodGet, r.Method)
	assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
	assert.Equal(t, "Rift", r.Header.Get("Apollographql-Client-Name"))
	assert.Equal(t, "1", r.Header.Get("Apollographql-Client-Version"))
	assert.Equal(t, "GetSeasonStage", r.URL.Query().Get("operationName"))

	var variables struct {
		Locale   string   `json:"hl"`
		StageIDs []string `json:"stageIds"`
	}

	err := json.Unmarshal([]byte(r.URL.Query().Get("variables")), &variables)
	require.NoError(t, err)
	assert.Equal(t, "en-US", variables.Locale)
	assert.Equal(t, []string{stageID}, variables.StageIDs)

	var extensions struct {
		PersistedQuery struct {
			Version int    `json:"version"`
			Hash    string `json:"sha256Hash"`
		} `json:"persistedQuery"`
	}

	err = json.Unmarshal([]byte(r.URL.Query().Get("extensions")), &extensions)
	require.NoError(t, err)
	assert.Equal(t, 1, extensions.PersistedQuery.Version)
	assert.Equal(
		t,
		"1b5a17d767e7b415b56d2319238468791089bc25cb782138697ae5eefbbea668",
		extensions.PersistedQuery.Hash,
	)
}
