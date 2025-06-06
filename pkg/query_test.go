package main

import (
	"github.com/grafana/grafana-plugin-sdk-go/data"
	_data "github.com/michelin/snowflake-grafana-datasource/pkg/data"
	"github.com/michelin/snowflake-grafana-datasource/pkg/query"
	sf "github.com/snowflakedb/gosnowflake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestIsTimeSeriesType_TrueWhenQueryTypeIsTimeSeries(t *testing.T) {
	qc := _data.QueryConfigStruct{QueryType: "time series"}
	assert.True(t, qc.IsTimeSeriesType())
}

func TestIsTimeSeriesType_FalseWhenQueryTypeIsNotTimeSeries(t *testing.T) {
	qc := _data.QueryConfigStruct{QueryType: "table"}
	assert.False(t, qc.IsTimeSeriesType())
}

func TestIsTimeSeriesType_FalseWhenQueryTypeIsEmpty(t *testing.T) {
	qc := _data.QueryConfigStruct{QueryType: ""}
	assert.False(t, qc.IsTimeSeriesType())
}

func TestFillTimesSeries_AppendsCorrectTimeValues(t *testing.T) {
	frame := data.NewFrame("")
	frame.Fields = append(frame.Fields, data.NewField("time", nil, []*time.Time{}))
	queryConfig := _data.QueryConfigStruct{
		Interval:  time.Minute,
		FillMode:  query.NullFill,
		QueryType: _data.TimeSeriesType,
	}
	fillTimesSeries(queryConfig, 0, 60000, 0, frame, 1, new(int), nil)
	assert.Equal(t, 1, frame.Fields[0].Len())
	assert.Equal(t, time.Unix(0, 0), *frame.Fields[0].At(0).(*time.Time))
}

func TestFillTimesSeries_AppendsFillValue(t *testing.T) {
	frame := data.NewFrame("")
	frame.Fields = append(frame.Fields, data.NewField("time", nil, []*time.Time{}))
	frame.Fields = append(frame.Fields, data.NewField("value", nil, []*float64{}))
	queryConfig := _data.QueryConfigStruct{
		Interval:  time.Minute,
		FillMode:  query.ValueFill,
		FillValue: 42.0,
		QueryType: _data.TimeSeriesType,
	}
	fillTimesSeries(queryConfig, 0, 60000, 0, frame, 2, new(int), nil)
	assert.Equal(t, 1, frame.Fields[1].Len())
	assert.Equal(t, 42.0, *frame.Fields[1].At(0).(*float64))
}

func TestFillTimesSeries_AppendsNilForNullFill(t *testing.T) {
	frame := data.NewFrame("")
	frame.Fields = append(frame.Fields, data.NewField("time", nil, []*time.Time{}))
	frame.Fields = append(frame.Fields, data.NewField("value", nil, []*float64{}))
	queryConfig := _data.QueryConfigStruct{
		Interval:  time.Minute,
		FillMode:  query.NullFill,
		QueryType: _data.TimeSeriesType,
	}
	fillTimesSeries(queryConfig, 0, 60000, 0, frame, 2, new(int), nil)
	assert.Equal(t, 1, frame.Fields[1].Len())
	assert.Nil(t, frame.Fields[1].At(0))
}

func TestFillTimesSeries_AppendsPreviousValue(t *testing.T) {
	frame := data.NewFrame("")
	frame.Fields = append(frame.Fields, data.NewField("time", nil, []*time.Time{}))
	frame.Fields = append(frame.Fields, data.NewField("value", nil, []*float64{}))
	queryConfig := _data.QueryConfigStruct{
		Interval:  time.Minute,
		FillMode:  query.PreviousFill,
		QueryType: _data.TimeSeriesType,
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
	queryConfig := _data.QueryConfigStruct{
		Interval:  time.Minute,
		FillMode:  query.NullFill,
		QueryType: "table",
	}
	fillTimesSeries(queryConfig, 0, 60000, 1, frame, 2, new(int), nil)
	assert.Equal(t, 0, frame.Fields[1].Len())
}

func TestAppendsNilWhenPreviousRowIsNil(t *testing.T) {
	frame := data.NewFrame("")
	frame.Fields = append(frame.Fields, data.NewField("time", nil, []*time.Time{}))
	frame.Fields = append(frame.Fields, data.NewField("value", nil, []*float64{}))
	queryConfig := _data.QueryConfigStruct{
		Interval:  time.Minute,
		FillMode:  query.PreviousFill,
		QueryType: _data.TimeSeriesType,
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
		getConnectionString(&config, _data.AuthenticationSecret{})
		require.Equal(t, 5, sf.MaxChunkDownloadWorkers)
	})

	t.Run("invalid MaxChunkDownloadWorkers", func(t *testing.T) {
		config.MaxChunkDownloadWorkers = "invalid"
		getConnectionString(&config, _data.AuthenticationSecret{})
		require.NotEqual(t, 5, sf.MaxChunkDownloadWorkers)
	})
}

func TestCustomJSONDecoderEnabled(t *testing.T) {
	config := pluginConfig{
		CustomJSONDecoderEnabled: true,
	}

	t.Run("CustomJSONDecoderEnabled true", func(t *testing.T) {
		getConnectionString(&config, _data.AuthenticationSecret{})
		require.True(t, sf.CustomJSONDecoderEnabled)
	})

	t.Run("CustomJSONDecoderEnabled false", func(t *testing.T) {
		config.CustomJSONDecoderEnabled = false
		getConnectionString(&config, _data.AuthenticationSecret{})
		require.False(t, sf.CustomJSONDecoderEnabled)
	})
}
