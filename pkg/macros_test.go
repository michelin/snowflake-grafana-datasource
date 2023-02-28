package main

import (
	"fmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestEvaluateMacro(t *testing.T) {

	timeRange := backend.TimeRange{
		From: time.Now(),
		To:   time.Now().Add(time.Minute),
	}

	configStruct := queryConfigStruct{
		TimeRange: timeRange,
	}

	tcs := []struct {
		args      []string
		name      string
		config    queryConfigStruct
		response  string
		err       string
		fillMode  string
		fillValue float64
	}{
		// __time
		{name: "__time", args: []string{"col"}, response: "TRY_TO_TIMESTAMP(col) AS time"},
		{name: "__time", args: []string{}, err: "missing time column argument for macro __time"},
		// __timeEpoch
		{name: "__timeEpoch", args: []string{}, err: "missing time column argument for macro __timeEpoch"},
		{name: "__timeEpoch", args: []string{"col"}, response: "extract(epoch from col) as time"},
		// __timeFilter
		{name: "__timeFilter", args: []string{}, err: "missing time column argument for macro __timeFilter"},
		{name: "__timeFilter", args: []string{"col"}, config: configStruct, response: "col BETWEEN '" + timeRange.From.UTC().Format(time.RFC3339Nano) + "' AND '" + timeRange.To.UTC().Format(time.RFC3339Nano) + "'"},
		{name: "__timeFilter", args: []string{"col", "'America/New_York'"}, config: configStruct, response: "CONVERT_TIMEZONE('UTC', 'America/New_York', col) >= '" + timeRange.From.UTC().Format(time.RFC3339Nano) + "' AND CONVERT_TIMEZONE('UTC', 'America/New_York', col) <= '" + timeRange.To.UTC().Format(time.RFC3339Nano) + "'"},
		// __timeFrom
		{name: "__timeFrom", args: []string{}, config: configStruct, response: "'" + timeRange.From.UTC().Format(time.RFC3339Nano) + "'"},
		// __timeTo
		{name: "__timeTo", args: []string{}, config: configStruct, response: "'" + timeRange.To.UTC().Format(time.RFC3339Nano) + "'"},
		// __timeGroup
		{name: "__timeGroup", args: []string{}, err: "macro __timeGroup needs time column and interval and optional fill value"},
		{name: "__timeGroup", args: []string{"col", "xxxx"}, err: "error parsing interval xxxx"},
		{name: "__timeGroup", args: []string{"col", "1d"}, response: "TIME_SLICE(TO_TIMESTAMP(col), 86400, 'SECOND', 'START')"},
		{name: "__timeGroup", args: []string{"col", "1d", "NULL"}, response: "TIME_SLICE(TO_TIMESTAMP(col), 86400, 'SECOND', 'START')", fillMode: NullFill},
		{name: "__timeGroup", args: []string{"col", "1d", "previous"}, response: "TIME_SLICE(TO_TIMESTAMP(col), 86400, 'SECOND', 'START')", fillMode: PreviousFill},
		{name: "__timeGroup", args: []string{"col", "1d", "12"}, response: "TIME_SLICE(TO_TIMESTAMP(col), 86400, 'SECOND', 'START')", fillMode: ValueFill, fillValue: 12},
		// __timeGroupAlias
		{name: "__timeGroupAlias", args: []string{}, err: "macro __timeGroup needs time column and interval and optional fill value"},
		{name: "__timeGroupAlias", args: []string{"col", "xxxx"}, err: "error parsing interval xxxx"},
		{name: "__timeGroupAlias", args: []string{"col", "1d"}, response: "TIME_SLICE(TO_TIMESTAMP(col), 86400, 'SECOND', 'START') AS time"},
		{name: "__timeGroupAlias", args: []string{"col", "1d", "NULL"}, response: "TIME_SLICE(TO_TIMESTAMP(col), 86400, 'SECOND', 'START') AS time", fillMode: NullFill},
		{name: "__timeGroupAlias", args: []string{"col", "1d", "previous"}, response: "TIME_SLICE(TO_TIMESTAMP(col), 86400, 'SECOND', 'START') AS time", fillMode: PreviousFill},
		{name: "__timeGroupAlias", args: []string{"col", "1d", "12"}, response: "TIME_SLICE(TO_TIMESTAMP(col), 86400, 'SECOND', 'START') AS time", fillMode: ValueFill, fillValue: 12},
		{name: "__timeGroupAlias", args: []string{"col", "1d", "test"}, err: "error parsing fill value test"},
		// __unixEpochFilter
		{name: "__unixEpochFilter", args: []string{}, err: "missing time column argument for macro __unixEpochFilter"},
		{name: "__unixEpochFilter", args: []string{"col"}, response: "col >= -62135596800 AND col <= -62135596800"},
		// __unixEpochNanoFilter
		{name: "__unixEpochNanoFilter", args: []string{}, err: "missing time column argument for macro __unixEpochNanoFilter"},
		{name: "__unixEpochNanoFilter", args: []string{"col"}, response: "col >= -6795364578871345152 AND col <= -6795364578871345152"},
		// __unixEpochNanoFrom
		{name: "__unixEpochNanoFrom", args: []string{}, response: "-6795364578871345152"},
		// __unixEpochNanoTo
		{name: "__unixEpochNanoTo", args: []string{}, response: "-6795364578871345152"},
		// __unixEpochGroup
		{name: "__unixEpochGroup", args: []string{}, err: "macro __unixEpochGroup needs time column and interval and optional fill value"},
		{name: "__unixEpochGroup", args: []string{"col", "xxxx"}, err: "error parsing interval xxxx"},
		{name: "__unixEpochGroup", args: []string{"col", "1d"}, response: "floor(col/86400)*86400"},
		{name: "__unixEpochGroup", args: []string{"col", "1d", "NULL"}, response: "floor(col/86400)*86400", fillMode: NullFill},
		{name: "__unixEpochGroup", args: []string{"col", "1d", "previous"}, response: "floor(col/86400)*86400", fillMode: PreviousFill},
		{name: "__unixEpochGroup", args: []string{"col", "1d", "12"}, response: "floor(col/86400)*86400", fillMode: ValueFill, fillValue: 12},
		{name: "__unixEpochGroup", args: []string{"col", "1d", "test"}, err: "error parsing fill value test"},
		// __unixEpochGroupAlias
		{name: "__unixEpochGroupAlias", args: []string{}, err: "macro __unixEpochGroup needs time column and interval and optional fill value"},
		{name: "__unixEpochGroupAlias", args: []string{"col", "xxxx"}, err: "error parsing interval xxxx"},
		{name: "__unixEpochGroupAlias", args: []string{"col", "1d"}, response: "floor(col/86400)*86400 AS time"},
		{name: "__unixEpochGroupAlias", args: []string{"col", "1d", "NULL"}, response: "floor(col/86400)*86400 AS time", fillMode: NullFill},
		{name: "__unixEpochGroupAlias", args: []string{"col", "1d", "previous"}, response: "floor(col/86400)*86400 AS time", fillMode: PreviousFill},
		{name: "__unixEpochGroupAlias", args: []string{"col", "1d", "12"}, response: "floor(col/86400)*86400 AS time", fillMode: ValueFill, fillValue: 12},
		{name: "__unixEpochGroupAlias", args: []string{"col", "1d", "test"}, err: "error parsing fill value test"},
		// default
		{name: "xxxx", args: []string{"col", "1d", "test"}, err: "unknown macro \"xxxx\""},
	}
	for i, tc := range tcs {
		t.Run(fmt.Sprintf("testcase for %s %d", tc.name, i), func(t *testing.T) {
			macro, err := evaluateMacro(tc.name, tc.args, &tc.config)
			if tc.err == "" {
				require.NoError(t, err, "input %s", tc.name)
				require.Equal(t, tc.response, macro)
				if tc.config.FillMode != "" {
					require.Equal(t, tc.fillMode, tc.config.FillMode)
				}
				if tc.config.FillValue != 0 {
					require.Equal(t, tc.fillValue, tc.config.FillValue)
				}
			} else {
				require.Error(t, err, "input %s", tc.name)
				require.Equal(t, tc.err, err.Error())
			}
		})
	}
}
