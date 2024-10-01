package main

import (
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsTimeSeriesType_TrueWhenQueryTypeIsTimeSeries(t *testing.T) {
	qc := queryConfigStruct{QueryType: "time series"}
	assert.True(t, qc.isTimeSeriesType())
}

func TestIsTimeSeriesType_FalseWhenQueryTypeIsNotTimeSeries(t *testing.T) {
	qc := queryConfigStruct{QueryType: "table"}
	assert.False(t, qc.isTimeSeriesType())
}

func TestIsTimeSeriesType_FalseWhenQueryTypeIsEmpty(t *testing.T) {
	qc := queryConfigStruct{QueryType: ""}
	assert.False(t, qc.isTimeSeriesType())
}

func TestMapFillMode(t *testing.T) {
	assert.Equal(t, data.FillModeValue, mapFillMode("value"))
	assert.Equal(t, data.FillModeNull, mapFillMode("null"))
	assert.Equal(t, data.FillModePrevious, mapFillMode("previous"))
	assert.Equal(t, data.FillModeNull, mapFillMode("unknown"))
	assert.Equal(t, data.FillModeNull, mapFillMode(""))
}
