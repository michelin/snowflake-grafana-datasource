package main

import (
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMapFillMode(t *testing.T) {
	assert.Equal(t, data.FillModeValue, mapFillMode("value"))
	assert.Equal(t, data.FillModeNull, mapFillMode("null"))
	assert.Equal(t, data.FillModePrevious, mapFillMode("previous"))
	assert.Equal(t, data.FillModeNull, mapFillMode("unknown"))
	assert.Equal(t, data.FillModeNull, mapFillMode(""))
}
