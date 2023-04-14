package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractPluginName(t *testing.T) {

	want := "myname"
	input := "/plugins/myname/metrics"
	got := extractPluginName(input)
	assert.Equal(t, got, want)
}
