package main

import (
	"fmt"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestContainsIgnoreCase(t *testing.T) {

	tcs := []struct {
		array   []string
		str     string
		success bool
	}{
		{array: []string{"value1", "value2"}, str: "VALUE1", success: true},
		{array: []string{"value1", "value2"}, str: "VALUE2", success: true},
		{array: []string{"value 1", "value2"}, str: "value 1", success: true},
		{array: []string{"value1", "value2"}, str: "", success: false},
		{array: []string{}, str: "", success: false},
	}
	for i, tc := range tcs {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			if tc.success {
				require.True(t, equalsIgnoreCase(tc.array, tc.str))
			} else {
				require.False(t, equalsIgnoreCase(tc.array, tc.str))
			}
		})
	}
}

func TestMin(t *testing.T) {

	tcs := []struct {
		val1   int64
		val2   int64
		result int64
	}{
		{val1: 1, val2: 2, result: 1},
		{val1: 2, val2: 2.0, result: 2},
		{val1: 2, val2: 1, result: 1},
	}
	for i, tc := range tcs {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			require.Equal(t, tc.result, Min(tc.val1, tc.val2))
		})
	}
}

func TestMax(t *testing.T) {

	tcs := []struct {
		val1   int64
		val2   int64
		result int64
	}{
		{val1: 1, val2: 2, result: 2},
		{val1: 2, val2: 2.0, result: 2},
		{val1: 2, val2: 1, result: 2},
	}
	for i, tc := range tcs {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			require.Equal(t, tc.result, Max(tc.val1, tc.val2))
		})
	}
}

func TestPreviousRowWithEmptyRows(t *testing.T) {
	rows := [][]interface{}{}
	result := previousRow(rows, 1)
	require.Nil(t, result)
}

func TestPreviousRowWithNonEmptyRowsAndIndexZero(t *testing.T) {
	rows := [][]interface{}{
		{"row1"},
		{"row2"},
	}
	result := previousRow(rows, 0)
	require.Equal(t, rows[0], result)
}

func TestPreviousRowWithNonEmptyRowsAndIndexGreaterThanZero(t *testing.T) {
	rows := [][]interface{}{
		{"row1"},
		{"row2"},
		{"row3"},
	}
	result := previousRow(rows, 2)
	require.Equal(t, rows[1], result)
}

func TestAppendsStringValueToFrameField(t *testing.T) {
	frame := data.NewFrame("test")
	frame.Fields = append(frame.Fields, data.NewField("field1", nil, []*string{}))
	value := "testString"
	insertFrameField(frame, value, 0)
	require.Equal(t, &value, frame.Fields[0].At(0))
}

func TestAppendsFloat64ValueToFrameField(t *testing.T) {
	frame := data.NewFrame("test")
	frame.Fields = append(frame.Fields, data.NewField("field1", nil, []*float64{}))
	value := float64(123.45)
	insertFrameField(frame, value, 0)
	require.Equal(t, &value, frame.Fields[0].At(0))
}

func TestAppendsInt64ValueToFrameField(t *testing.T) {
	frame := data.NewFrame("test")
	frame.Fields = append(frame.Fields, data.NewField("field1", nil, []*int64{}))
	value := int64(123)
	insertFrameField(frame, value, 0)
	require.Equal(t, &value, frame.Fields[0].At(0))
}

func TestAppendsBoolValueToFrameField(t *testing.T) {
	frame := data.NewFrame("test")
	frame.Fields = append(frame.Fields, data.NewField("field1", nil, []*bool{}))
	value := true
	insertFrameField(frame, value, 0)
	require.Equal(t, &value, frame.Fields[0].At(0))
}

func TestAppendsTimeValueToFrameField(t *testing.T) {
	frame := data.NewFrame("test")
	frame.Fields = append(frame.Fields, data.NewField("field1", nil, []*time.Time{}))
	value := time.Now()
	insertFrameField(frame, value, 0)
	require.Equal(t, &value, frame.Fields[0].At(0))
}
