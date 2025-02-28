package query

import (
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMapFillMode(t *testing.T) {
	assert.Equal(t, data.FillModeValue, MapFillMode("value"))
	assert.Equal(t, data.FillModeNull, MapFillMode("null"))
	assert.Equal(t, data.FillModePrevious, MapFillMode("previous"))
	assert.Equal(t, data.FillModeNull, MapFillMode("unknown"))
	assert.Equal(t, data.FillModeNull, MapFillMode(""))
}
