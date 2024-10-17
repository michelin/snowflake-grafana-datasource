package main

import (
	"github.com/grafana/grafana-plugin-sdk-go/data"
	sf "github.com/snowflakedb/gosnowflake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
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

// Helper functions to create pointers
func timePtr(t time.Time) *time.Time {
	return &t
}

func float64Ptr(f float64) *float64 {
	return &f
}

func TestMapFillMode(t *testing.T) {
	assert.Equal(t, data.FillModeValue, mapFillMode("value"))
	assert.Equal(t, data.FillModeNull, mapFillMode("null"))
	assert.Equal(t, data.FillModePrevious, mapFillMode("previous"))
	assert.Equal(t, data.FillModeNull, mapFillMode("unknown"))
	assert.Equal(t, data.FillModeNull, mapFillMode(""))
}

func TestFillTimesSeries_AppendsCorrectTimeValues(t *testing.T) {
	frame := data.NewFrame("")
	frame.Fields = append(frame.Fields, data.NewField("time", nil, []*time.Time{}))
	queryConfig := queryConfigStruct{
		Interval:  time.Minute,
		FillMode:  NullFill,
		QueryType: timeSeriesType,
	}
	fillTimesSeries(queryConfig, 0, 60000, 0, frame, 1, new(int), nil)
	assert.Equal(t, 1, frame.Fields[0].Len())
	assert.Equal(t, time.Unix(0, 0), *frame.Fields[0].At(0).(*time.Time))
}

func TestFillTimesSeries_AppendsFillValue(t *testing.T) {
	frame := data.NewFrame("")
	frame.Fields = append(frame.Fields, data.NewField("time", nil, []*time.Time{}))
	frame.Fields = append(frame.Fields, data.NewField("value", nil, []*float64{}))
	queryConfig := queryConfigStruct{
		Interval:  time.Minute,
		FillMode:  ValueFill,
		FillValue: 42.0,
		QueryType: timeSeriesType,
	}
	fillTimesSeries(queryConfig, 0, 60000, 0, frame, 2, new(int), nil)
	assert.Equal(t, 1, frame.Fields[1].Len())
	assert.Equal(t, 42.0, *frame.Fields[1].At(0).(*float64))
}

func TestFillTimesSeries_AppendsNilForNullFill(t *testing.T) {
	frame := data.NewFrame("")
	frame.Fields = append(frame.Fields, data.NewField("time", nil, []*time.Time{}))
	frame.Fields = append(frame.Fields, data.NewField("value", nil, []*float64{}))
	queryConfig := queryConfigStruct{
		Interval:  time.Minute,
		FillMode:  NullFill,
		QueryType: timeSeriesType,
	}
	fillTimesSeries(queryConfig, 0, 60000, 0, frame, 2, new(int), nil)
	assert.Equal(t, 1, frame.Fields[1].Len())
	assert.Nil(t, frame.Fields[1].At(0))
}

func TestFillTimesSeries_AppendsPreviousValue(t *testing.T) {
	frame := data.NewFrame("")
	frame.Fields = append(frame.Fields, data.NewField("time", nil, []*time.Time{}))
	frame.Fields = append(frame.Fields, data.NewField("value", nil, []*float64{}))
	queryConfig := queryConfigStruct{
		Interval:  time.Minute,
		FillMode:  PreviousFill,
		QueryType: timeSeriesType,
	}
	previousRow := []interface{}{time.Unix(0, 0), 42.0}
	fillTimesSeries(queryConfig, 0, 60000, 0, frame, 2, new(int), previousRow)
	assert.Equal(t, 1, frame.Fields[1].Len())
	assert.Equal(t, 42.0, *frame.Fields[1].At(0).(*float64))
}

func TestFillTimesSeries_DoesNotAppendWhenNotTimeSeries(t *testing.T) {
	frame := data.NewFrame("")
	frame.Fields = append(frame.Fields, data.NewField("time", nil, []*time.Time{}))
	frame.Fields = append(frame.Fields, data.NewField("value", nil, []*float64{}))
	queryConfig := queryConfigStruct{
		Interval:  time.Minute,
		FillMode:  NullFill,
		QueryType: "table",
	}
	fillTimesSeries(queryConfig, 0, 60000, 1, frame, 2, new(int), nil)
	assert.Equal(t, 0, frame.Fields[1].Len())
}

func TestAppendsNilWhenPreviousRowIsNil(t *testing.T) {
	frame := data.NewFrame("")
	frame.Fields = append(frame.Fields, data.NewField("time", nil, []*time.Time{}))
	frame.Fields = append(frame.Fields, data.NewField("value", nil, []*float64{}))
	queryConfig := queryConfigStruct{
		Interval:  time.Minute,
		FillMode:  PreviousFill,
		QueryType: timeSeriesType,
	}
	fillTimesSeries(queryConfig, 0, 60000, 0, frame, 2, new(int), nil)
	assert.Equal(t, 1, frame.Fields[1].Len())
	assert.Nil(t, frame.Fields[1].At(0))
}

func TestMaxChunkDownloadWorkers(t *testing.T) {
	config := pluginConfig{
		MaxChunkDownloadWorkers: "5",
	}

	t.Run("valid MaxChunkDownloadWorkers", func(t *testing.T) {
		getConnectionString(&config, "", "")
		require.Equal(t, 5, sf.MaxChunkDownloadWorkers)
	})

	t.Run("invalid MaxChunkDownloadWorkers", func(t *testing.T) {
		config.MaxChunkDownloadWorkers = "invalid"
		getConnectionString(&config, "", "")
		require.NotEqual(t, 5, sf.MaxChunkDownloadWorkers)
	})
}

func TestCustomJSONDecoderEnabled(t *testing.T) {
	config := pluginConfig{
		CustomJSONDecoderEnabled: true,
	}

	t.Run("CustomJSONDecoderEnabled true", func(t *testing.T) {
		getConnectionString(&config, "", "")
		require.True(t, sf.CustomJSONDecoderEnabled)
	})

	t.Run("CustomJSONDecoderEnabled false", func(t *testing.T) {
		config.CustomJSONDecoderEnabled = false
		getConnectionString(&config, "", "")
		require.False(t, sf.CustomJSONDecoderEnabled)
	})
}
