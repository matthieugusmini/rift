package ui //nolint:testpackage // White-box tests verify league title formatting.

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLeagueItemTitleIncludesCurrentRegionalFlags(t *testing.T) {
	tests := []struct {
		leagueName string
		want       string
	}{
		{leagueName: "LCS", want: "LCS • 🇺🇸"},
		{leagueName: "CBLOL", want: "CBLOL • 🇧🇷"},
	}

	for _, test := range tests {
		t.Run(test.leagueName, func(t *testing.T) {
			item := leagueItem{leagueName: test.leagueName}

			assert.Equal(t, test.want, item.Title())
		})
	}
}
