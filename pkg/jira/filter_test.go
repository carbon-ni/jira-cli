package jira

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterCollectionGet(t *testing.T) {
	cltn := FilterCollection{NewNumCommentsFilter(5)}
	assert.Equal(t, uint(5), cltn.Get(cltn[0].Key()))
	assert.Nil(t, cltn.Get("unknown"))
}

func TestFilterCollectionGetInt(t *testing.T) {
	cltn := FilterCollection{NewNumCommentsFilter(5)}
	assert.Equal(t, 5, cltn.GetInt(cltn[0].Key()))
	assert.Equal(t, 0, cltn.GetInt("unknown"))
}
