package jira

// Filter groups filterable config.
type Filter interface {
	Key() FilterKey
	Val() interface{}
}

// FilterKey represents filter key for issue.
type FilterKey string

// FilterCollection is a group of unique filters.
type FilterCollection []Filter

// Get returns filter value as it is passed.
func (flt FilterCollection) Get(key FilterKey) interface{} {
	for _, f := range flt {
		if f.Key() == key {
			return f.Val()
		}
	}
	return nil
}

// GetInt returns filter value as an integer.
func (flt FilterCollection) GetInt(key FilterKey) int {
	for _, f := range flt {
		if f.Key() != key {
			continue
		}
		if v, ok := f.Val().(uint); ok {
			return int(v)
		}
	}
	return 0
}

// KeyIssueNumComments is a filter key for issue comments.
const KeyIssueNumComments = FilterKey("issue-num-comments")

// NumCommentsFilter is a filter for issue comments.
type NumCommentsFilter struct {
	key   FilterKey
	value uint
}

// NewNumCommentsFilter constructs a filter to limit number of comments.
func NewNumCommentsFilter(value uint) NumCommentsFilter {
	return NumCommentsFilter{
		key:   KeyIssueNumComments,
		value: value,
	}
}

// Key returns key of this filter.
func (ncf NumCommentsFilter) Key() FilterKey {
	return ncf.key
}

// Val returns value of this filter.
func (ncf NumCommentsFilter) Val() interface{} {
	return ncf.value
}
