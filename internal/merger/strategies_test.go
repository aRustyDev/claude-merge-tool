package merger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeStrategy_IsValid(t *testing.T) {
	tests := []struct {
		strategy MergeStrategy
		valid    bool
	}{
		{StrategyReplace, true},
		{StrategyAppend, true},
		{StrategyPrepend, true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.strategy), func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.strategy.IsValid())
		})
	}
}

func TestApplyStrategy(t *testing.T) {
	tests := []struct {
		name     string
		strategy MergeStrategy
		old      string
		new      string
		expected string
	}{
		{"replace", StrategyReplace, "old content", "new content", "new content"},
		{"append", StrategyAppend, "old content", "new content", "old content\nnew content"},
		{"prepend", StrategyPrepend, "old content", "new content", "new content\nold content"},
		{"append_empty_old", StrategyAppend, "", "new content", "new content"},
		{"prepend_empty_old", StrategyPrepend, "", "new content", "new content"},
		{"replace_empty_old", StrategyReplace, "", "new content", "new content"},
		{"replace_empty_new", StrategyReplace, "old content", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyStrategy(tt.strategy, tt.old, tt.new)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestApplyStrategy_InvalidStrategy(t *testing.T) {
	// Invalid strategy should default to replace
	result := ApplyStrategy("invalid", "old", "new")
	assert.Equal(t, "new", result)
}

func TestMergeStrategy_String(t *testing.T) {
	tests := []struct {
		strategy MergeStrategy
		expected string
	}{
		{StrategyReplace, "replace"},
		{StrategyAppend, "append"},
		{StrategyPrepend, "prepend"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.strategy))
		})
	}
}