package ui //nolint:testpackage // White-box tests exercise the internal color palette.

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewThemeSelectsTerminalBackgroundVariant(t *testing.T) {
	lightTheme := newTheme(false)
	darkTheme := newTheme(true)

	assert.False(t, lightTheme.isDark)
	assert.True(t, darkTheme.isDark)
	assert.NotEqual(t, lightTheme.textPrimary, darkTheme.textPrimary)
	assert.NotEqual(t, lightTheme.selected, darkTheme.selected)
}
