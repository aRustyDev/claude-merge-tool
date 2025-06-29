package merger

import (
	"testing"

	"github.com/arustydev/claude-merge/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPriorityMerger_MergeAll_ExplicitPriority(t *testing.T) {
	// Test explicit priority overrides
	config1 := &config.Config{
		Metadata: config.Metadata{
			Title:    "Base Config",
			Priority: config.NewExplicitPriority(5),
		},
		Sections: map[string]config.Section{
			"section1": {
				Content:  "Base content",
				Priority: config.NewExplicitPriority(5),
			},
		},
	}

	config2 := &config.Config{
		Metadata: config.Metadata{
			Title:    "Override Config",
			Priority: config.NewExplicitPriority(10),
		},
		Sections: map[string]config.Section{
			"section1": {
				Content:  "Override content",
				Priority: config.NewExplicitPriority(10),
			},
		},
	}

	merger := NewPriorityMerger(false)
	result, err := merger.MergeAll([]*config.Config{config1, config2})
	require.NoError(t, err)

	assert.Equal(t, "Override Config", result.Metadata.Title)
	assert.Equal(t, "Override content", result.Sections["section1"].Content)
}

func TestPriorityMerger_MergeAll_RelativePriority(t *testing.T) {
	config1 := &config.Config{
		Sections: map[string]config.Section{
			"section1": {
				Content:  "Relative content",
				Priority: config.NewRelativePriority(10),
			},
		},
	}

	config2 := &config.Config{
		Sections: map[string]config.Section{
			"section1": {
				Content:  "Should not override",
				Priority: config.NewRelativePriority(5),
			},
		},
	}

	merger := NewPriorityMerger(false)
	result, err := merger.MergeAll([]*config.Config{config1, config2})
	require.NoError(t, err)

	assert.Equal(t, "Relative content", result.Sections["section1"].Content)
}

func TestPriorityMerger_MergeAll_FileOrder(t *testing.T) {
	config1 := &config.Config{
		Sections: map[string]config.Section{
			"section1": {Content: "First content"},
		},
	}

	config2 := &config.Config{
		Sections: map[string]config.Section{
			"section1": {Content: "Second content"},
		},
	}

	merger := NewPriorityMerger(false)
	result, err := merger.MergeAll([]*config.Config{config1, config2})
	require.NoError(t, err)

	// Second config should win with file order precedence
	assert.Equal(t, "Second content", result.Sections["section1"].Content)
}

func TestPriorityMerger_MergeAll_MergePoints(t *testing.T) {
	config1 := &config.Config{
		MergePoints: map[string]config.MergePoint{
			"test-point": {
				Placeholder: "<!-- MERGE:test -->",
				Default:     "default content",
				Priority:    config.NewRelativePriority(5),
			},
		},
	}

	config2 := &config.Config{
		MergePoints: map[string]config.MergePoint{
			"test-point": {
				Placeholder: "<!-- MERGE:test -->",
				Default:     "override content",
				Priority:    config.NewExplicitPriority(10),
			},
		},
	}

	merger := NewPriorityMerger(false)
	result, err := merger.MergeAll([]*config.Config{config1, config2})
	require.NoError(t, err)

	assert.Equal(t, "override content", result.MergePoints["test-point"].Default)
}

func TestPriorityMerger_MergeAll_MergeTargets(t *testing.T) {
	config1 := &config.Config{
		MergeTargets: map[string]config.MergeTarget{
			"test-target": {
				Strategy: "replace",
				Content:  "base target",
				Priority: config.NewRelativePriority(5),
			},
		},
	}

	config2 := &config.Config{
		MergeTargets: map[string]config.MergeTarget{
			"test-target": {
				Strategy: "append",
				Content:  "override target",
				Priority: config.NewExplicitPriority(10),
			},
		},
	}

	merger := NewPriorityMerger(false)
	result, err := merger.MergeAll([]*config.Config{config1, config2})
	require.NoError(t, err)

	assert.Equal(t, "append", result.MergeTargets["test-target"].Strategy)
	assert.Equal(t, "override target", result.MergeTargets["test-target"].Content)
}

func TestPriorityMerger_MergeAll_EmptyConfigs(t *testing.T) {
	merger := NewPriorityMerger(false)
	_, err := merger.MergeAll([]*config.Config{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no configurations to merge")
}

func TestPriorityMerger_MergeAll_SingleConfig(t *testing.T) {
	config1 := &config.Config{
		Metadata: config.Metadata{Title: "Single Config"},
		Sections: map[string]config.Section{
			"section1": {Content: "Single content"},
		},
	}

	merger := NewPriorityMerger(false)
	result, err := merger.MergeAll([]*config.Config{config1})
	require.NoError(t, err)

	assert.Equal(t, "Single Config", result.Metadata.Title)
	assert.Equal(t, "Single content", result.Sections["section1"].Content)
}

func TestPriorityMerger_MergeAll_AdditiveSections(t *testing.T) {
	config1 := &config.Config{
		Sections: map[string]config.Section{
			"section1": {Content: "Content 1"},
		},
	}

	config2 := &config.Config{
		Sections: map[string]config.Section{
			"section2": {Content: "Content 2"},
		},
	}

	merger := NewPriorityMerger(false)
	result, err := merger.MergeAll([]*config.Config{config1, config2})
	require.NoError(t, err)

	assert.Equal(t, "Content 1", result.Sections["section1"].Content)
	assert.Equal(t, "Content 2", result.Sections["section2"].Content)
}

func TestPriorityMerger_MergeAll_ExplicitBeatsRelative(t *testing.T) {
	config1 := &config.Config{
		Sections: map[string]config.Section{
			"section1": {
				Content:  "Relative high value",
				Priority: config.NewRelativePriority(100),
			},
		},
	}

	config2 := &config.Config{
		Sections: map[string]config.Section{
			"section1": {
				Content:  "Explicit low value",
				Priority: config.NewExplicitPriority(1),
			},
		},
	}

	merger := NewPriorityMerger(false)
	result, err := merger.MergeAll([]*config.Config{config1, config2})
	require.NoError(t, err)

	// Explicit priority should always beat relative, regardless of values
	assert.Equal(t, "Explicit low value", result.Sections["section1"].Content)
}

func TestPriorityMerger_MergeAll_MetadataFields(t *testing.T) {
	config1 := &config.Config{
		Metadata: config.Metadata{
			Title:       "Base Title",
			Description: "Base Description",
			Version:     "1.0.0",
			Priority:    config.NewRelativePriority(5),
		},
	}

	config2 := &config.Config{
		Metadata: config.Metadata{
			Title:    "Override Title",
			Language: "golang",
			Priority: config.NewExplicitPriority(10),
		},
	}

	merger := NewPriorityMerger(false)
	result, err := merger.MergeAll([]*config.Config{config1, config2})
	require.NoError(t, err)

	assert.Equal(t, "Override Title", result.Metadata.Title)
	assert.Equal(t, "Base Description", result.Metadata.Description) // Should keep from config1
	assert.Equal(t, "1.0.0", result.Metadata.Version)               // Should keep from config1
	assert.Equal(t, "golang", result.Metadata.Language)             // Should get from config2
}

func TestNewPriorityMerger(t *testing.T) {
	merger := NewPriorityMerger(true)
	assert.NotNil(t, merger)
	assert.True(t, merger.debug)

	merger2 := NewPriorityMerger(false)
	assert.NotNil(t, merger2)
	assert.False(t, merger2.debug)
}