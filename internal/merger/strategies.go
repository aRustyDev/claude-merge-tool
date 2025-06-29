package merger

// MergeStrategy defines how content should be combined
type MergeStrategy string

const (
	// StrategyReplace means the new content completely replaces the old
	StrategyReplace MergeStrategy = "replace"

	// StrategyAppend means the new content is added after the old
	StrategyAppend MergeStrategy = "append"

	// StrategyPrepend means the new content is added before the old
	StrategyPrepend MergeStrategy = "prepend"
)

// IsValid checks if a strategy string is valid
func (s MergeStrategy) IsValid() bool {
	switch s {
	case StrategyReplace, StrategyAppend, StrategyPrepend:
		return true
	default:
		return false
	}
}

// ApplyStrategy applies a merge strategy to combine old and new content
func ApplyStrategy(strategy MergeStrategy, old, new string) string {
	switch strategy {
	case StrategyReplace:
		return new
	case StrategyAppend:
		if old == "" {
			return new
		}
		return old + "\n" + new
	case StrategyPrepend:
		if old == "" {
			return new
		}
		return new + "\n" + old
	default:
		return new // Default to replace
	}
}