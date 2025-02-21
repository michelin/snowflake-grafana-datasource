package query

import (
	"fmt"
	"github.com/michelin/snowflake-grafana-datasource/pkg/data"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/stretchr/testify/require"
)

func TestEvaluateMacro(t *testing.T) {

	timeRange := backend.TimeRange{
		From: time.Now(),
		To:   time.Now().Add(time.Minute),
	}

	configStruct := data.QueryConfigStruct{
		TimeRange: timeRange,
	}

	tcs := []struct {
		args      []string
		name      string
		config    data.QueryConfigStruct
		response  string
		err       string
		fillMode  string
		fillValue float64
	}{
		// __time
		{name: "__time", args: []string{}, err: "missing time column argument for macro __time"},
		{name: "__time", args: []string{""}, err: "missing time column argument for macro __time"},
		{name: "__time", args: []string{"col"}, response: "TRY_TO_TIMESTAMP_NTZ(col) AS time"},
		// __timeEpoch
		{name: "__timeEpoch", args: []string{}, err: "missing time column argument for macro __timeEpoch"},
		{name: "__timeEpoch", args: []string{""}, err: "missing time column argument for macro __timeEpoch"},
		{name: "__timeEpoch", args: []string{"col"}, response: "extract(epoch from col) as time"},
		// __timeFilter
		{name: "__timeFilter", args: []string{}, err: "missing time column argument for macro __timeFilter"},
		{name: "__timeFilter", args: []string{""}, err: "missing time column argument for macro __timeFilter"},
		{name: "__timeFilter", args: []string{"col"}, config: configStruct, response: "col > CONVERT_TIMEZONE('UTC', 'UTC', '" + timeRange.From.UTC().Format(time.RFC3339Nano) + "'::timestamp_ntz) AND col < CONVERT_TIMEZONE('UTC', 'UTC', '" + timeRange.To.UTC().Format(time.RFC3339Nano) + "'::timestamp_ntz)"},
		{name: "__timeFilter", args: []string{"col", "'America/New_York'"}, config: configStruct, response: "col > CONVERT_TIMEZONE('UTC', 'America/New_York', '" + timeRange.From.UTC().Format(time.RFC3339Nano) + "'::timestamp_ntz) AND col < CONVERT_TIMEZONE('UTC', 'America/New_York', '" + timeRange.To.UTC().Format(time.RFC3339Nano) + "'::timestamp_ntz)"},
		// __timeTzFilter
		{name: "__timeTzFilter", args: []string{}, err: "missing time column argument for macro __timeTzFilter"},
		{name: "__timeTzFilter", args: []string{""}, err: "missing time column argument for macro __timeTzFilter"},
		{name: "__timeTzFilter", args: []string{"col"}, config: configStruct, response: "col > '" + timeRange.From.UTC().Format(time.RFC3339Nano) + "'::timestamp_tz AND col < '" + timeRange.To.UTC().Format(time.RFC3339Nano) + "'::timestamp_tz"},
		// __timeFrom
		{name: "__timeFrom", args: []string{}, config: configStruct, response: "'" + timeRange.From.UTC().Format(time.RFC3339Nano) + "'"},
		// __timeTo
		{name: "__timeTo", args: []string{}, config: configStruct, response: "'" + timeRange.To.UTC().Format(time.RFC3339Nano) + "'"},
		// __timeGroup
		{name: "__timeGroup", args: []string{}, err: "macro __timeGroup needs time column and interval and optional fill value"},
		{name: "__timeGroup", args: []string{"col", "xxxx"}, err: "error parsing interval xxxx"},
		{name: "__timeGroup", args: []string{"col", "1d"}, response: "TIME_SLICE(TO_TIMESTAMP_NTZ(col), 86400, 'SECOND', 'START')"},
		{name: "__timeGroup", args: []string{"col", "500ms"}, response: "TIME_SLICE(TO_TIMESTAMP_NTZ(col), 1, 'SECOND', 'START')"},
		{name: "__timeGroup", args: []string{"col", "1d", "NULL"}, response: "TIME_SLICE(TO_TIMESTAMP_NTZ(col), 86400, 'SECOND', 'START')", fillMode: NullFill},
		{name: "__timeGroup", args: []string{"col", "1d", "previous"}, response: "TIME_SLICE(TO_TIMESTAMP_NTZ(col), 86400, 'SECOND', 'START')", fillMode: PreviousFill},
		{name: "__timeGroup", args: []string{"col", "1d", "12"}, response: "TIME_SLICE(TO_TIMESTAMP_NTZ(col), 86400, 'SECOND', 'START')", fillMode: ValueFill, fillValue: 12},
		{name: "__timeGroup", args: []string{"col", "7d"}, response: "TIME_SLICE(TO_TIMESTAMP_NTZ(col), 1, 'WEEK', 'START')"},
		{name: "__timeGroup", args: []string{"col", "2w"}, response: "TIME_SLICE(TO_TIMESTAMP_NTZ(col), 2, 'WEEK', 'START')"},
		// __timeGroupAlias
		{name: "__timeGroupAlias", args: []string{}, err: "macro __timeGroup needs time column and interval and optional fill value"},
		{name: "__timeGroupAlias", args: []string{"col", "xxxx"}, err: "error parsing interval xxxx"},
		{name: "__timeGroupAlias", args: []string{"col", "1d"}, response: "TIME_SLICE(TO_TIMESTAMP_NTZ(col), 86400, 'SECOND', 'START') AS time"},
		{name: "__timeGroupAlias", args: []string{"col", "1d", "NULL"}, response: "TIME_SLICE(TO_TIMESTAMP_NTZ(col), 86400, 'SECOND', 'START') AS time", fillMode: NullFill},
		{name: "__timeGroupAlias", args: []string{"col", "1d", "previous"}, response: "TIME_SLICE(TO_TIMESTAMP_NTZ(col), 86400, 'SECOND', 'START') AS time", fillMode: PreviousFill},
		{name: "__timeGroupAlias", args: []string{"col", "1d", "12"}, response: "TIME_SLICE(TO_TIMESTAMP_NTZ(col), 86400, 'SECOND', 'START') AS time", fillMode: ValueFill, fillValue: 12},
		{name: "__timeGroupAlias", args: []string{"col", "1d", "test"}, err: "error parsing fill value test"},
		// __unixEpochFilter
		{name: "__unixEpochFilter", args: []string{}, err: "missing time column argument for macro __unixEpochFilter"},
		{name: "__unixEpochFilter", args: []string{""}, err: "missing time column argument for macro __unixEpochFilter"},
		{name: "__unixEpochFilter", args: []string{"col"}, response: "col >= -62135596800 AND col <= -62135596800"},
		// __unixEpochNanoFilter
		{name: "__unixEpochNanoFilter", args: []string{}, err: "missing time column argument for macro __unixEpochNanoFilter"},
		{name: "__unixEpochNanoFilter", args: []string{""}, err: "missing time column argument for macro __unixEpochNanoFilter"},
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
		// __timeRoundTo
		{name: "__timeRoundTo", args: []string{"test"}, err: "macro __timeRoundTo first argument must be a integer"},
		{name: "__timeRoundTo", args: []string{"-1"}, err: "macro __timeRoundTo first argument must be a positive Integer"},
		{name: "__timeRoundTo", args: []string{"-1", ""}, err: "macro __timeRoundTo only 1 argument allowed"},
		// __timeRoundFrom
		{name: "__timeRoundFrom", args: []string{"test"}, err: "macro __timeRoundFrom first argument must be a integer"},
		{name: "__timeRoundFrom", args: []string{"-1"}, err: "macro __timeRoundFrom first argument must be a positive Integer"},
		{name: "__timeRoundFrom", args: []string{"-1", ""}, err: "macro __timeRoundFrom only 1 argument allowed"},
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

func TestInterpolate(t *testing.T) {
	tests := []struct {
		name          string
		configStruct  *data.QueryConfigStruct
		expectedSQL   string
		expectedError string
	}{
		{
			name: "valid macro",
			configStruct: &data.QueryConfigStruct{
				RawQuery: "SELECT * FROM table WHERE $__timeFilter(col)",
				TimeRange: backend.TimeRange{
					From: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
					To:   time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
				},
			},
			expectedSQL:   "SELECT * FROM table WHERE col > CONVERT_TIMEZONE('UTC', 'UTC', '2020-01-01T00:00:00Z'::timestamp_ntz) AND col < CONVERT_TIMEZONE('UTC', 'UTC', '2020-01-02T00:00:00Z'::timestamp_ntz)",
			expectedError: "",
		},
		{
			name: "missing time column argument",
			configStruct: &data.QueryConfigStruct{
				RawQuery: "SELECT * FROM table WHERE $__timeFilter()",
			},
			expectedSQL:   "",
			expectedError: "missing time column argument for macro __timeFilter",
		},
		{
			name: "valid snowflake system macro",
			configStruct: &data.QueryConfigStruct{
				RawQuery: "SELECT SYSTEM$TYPEOF('a')",
			},
			expectedSQL:   "SELECT SYSTEM$TYPEOF('a')",
			expectedError: "",
		},
		{
			name: "unknown macro",
			configStruct: &data.QueryConfigStruct{
				RawQuery: "SELECT * FROM table WHERE $__unknownMacro(col)",
			},
			expectedSQL:   "",
			expectedError: "unknown macro \"__unknownMacro\"",
		},
		{
			name: "check __timeRoundTo with default 15min",
			configStruct: &data.QueryConfigStruct{
				RawQuery: "SELECT * FROM table WHERE $__timeRoundTo()",
				TimeRange: backend.TimeRange{
					From: time.Date(2020, 1, 1, 0, 7, 0, 0, time.UTC),
					To:   time.Date(2020, 1, 2, 0, 7, 0, 0, time.UTC),
				},
			},
			expectedSQL:   "SELECT * FROM table WHERE '2020-01-02T00:15:00Z'",
			expectedError: "",
		},
		{
			name: "check __timeRoundFrom with default 15min",
			configStruct: &data.QueryConfigStruct{
				RawQuery: "SELECT * FROM table WHERE $__timeRoundFrom()",
				TimeRange: backend.TimeRange{
					From: time.Date(2020, 1, 1, 0, 7, 0, 0, time.UTC),
					To:   time.Date(2020, 1, 2, 0, 7, 0, 0, time.UTC),
				},
			},
			expectedSQL:   "SELECT * FROM table WHERE '2020-01-01T00:00:00Z'",
			expectedError: "",
		},
		{
			name: "check __timeRoundTo with 5min",
			configStruct: &data.QueryConfigStruct{
				RawQuery: "SELECT * FROM table WHERE $__timeRoundTo(5)",
				TimeRange: backend.TimeRange{
					From: time.Date(2020, 1, 1, 0, 7, 0, 0, time.UTC),
					To:   time.Date(2020, 1, 2, 0, 7, 0, 0, time.UTC),
				},
			},
			expectedSQL:   "SELECT * FROM table WHERE '2020-01-02T00:10:00Z'",
			expectedError: "",
		},
		{
			name: "check __timeRoundFrom with 5min",
			configStruct: &data.QueryConfigStruct{
				RawQuery: "SELECT * FROM table WHERE $__timeRoundFrom(5)",
				TimeRange: backend.TimeRange{
					From: time.Date(2020, 1, 1, 0, 7, 0, 0, time.UTC),
					To:   time.Date(2020, 1, 2, 0, 7, 0, 0, time.UTC),
				},
			},
			expectedSQL:   "SELECT * FROM table WHERE '2020-01-01T00:05:00Z'",
			expectedError: "",
		},
		{
			name: "check __timeRoundTo with 10min",
			configStruct: &data.QueryConfigStruct{
				RawQuery: "SELECT * FROM table WHERE $__timeRoundTo(10)",
				TimeRange: backend.TimeRange{
					From: time.Date(2020, 1, 1, 0, 7, 0, 0, time.UTC),
					To:   time.Date(2020, 1, 2, 0, 7, 0, 0, time.UTC),
				},
			},
			expectedSQL:   "SELECT * FROM table WHERE '2020-01-02T00:10:00Z'",
			expectedError: "",
		},
		{
			name: "check __timeRoundFrom with 10min",
			configStruct: &data.QueryConfigStruct{
				RawQuery: "SELECT * FROM table WHERE $__timeRoundFrom(10)",
				TimeRange: backend.TimeRange{
					From: time.Date(2020, 1, 1, 0, 7, 0, 0, time.UTC),
					To:   time.Date(2020, 1, 2, 0, 7, 0, 0, time.UTC),
				},
			},
			expectedSQL:   "SELECT * FROM table WHERE '2020-01-01T00:00:00Z'",
			expectedError: "",
		},
		{
			name: "check __timeRoundTo with 30min",
			configStruct: &data.QueryConfigStruct{
				RawQuery: "SELECT * FROM table WHERE $__timeRoundTo(30)",
				TimeRange: backend.TimeRange{
					From: time.Date(2020, 1, 1, 0, 7, 0, 0, time.UTC),
					To:   time.Date(2020, 1, 2, 0, 7, 0, 0, time.UTC),
				},
			},
			expectedSQL:   "SELECT * FROM table WHERE '2020-01-02T00:30:00Z'",
			expectedError: "",
		},
		{
			name: "check __timeRoundFrom with 30min",
			configStruct: &data.QueryConfigStruct{
				RawQuery: "SELECT * FROM table WHERE $__timeRoundFrom(30)",
				TimeRange: backend.TimeRange{
					From: time.Date(2020, 1, 1, 0, 59, 0, 0, time.UTC),
					To:   time.Date(2020, 1, 2, 0, 7, 0, 0, time.UTC),
				},
			},
			expectedSQL:   "SELECT * FROM table WHERE '2020-01-01T00:30:00Z'",
			expectedError: "",
		},
		{
			name: "check __timeRoundTo with 1440min",
			configStruct: &data.QueryConfigStruct{
				RawQuery: "SELECT * FROM table WHERE $__timeRoundTo(1440)",
				TimeRange: backend.TimeRange{
					From: time.Date(2020, 1, 1, 0, 7, 0, 0, time.UTC),
					To:   time.Date(2020, 1, 2, 0, 7, 0, 0, time.UTC),
				},
			},
			expectedSQL:   "SELECT * FROM table WHERE '2020-01-03T00:00:00Z'",
			expectedError: "",
		},
		{
			name: "check __timeRoundFrom with 1440min",
			configStruct: &data.QueryConfigStruct{
				RawQuery: "SELECT * FROM table WHERE $__timeRoundFrom(1440)",
				TimeRange: backend.TimeRange{
					From: time.Date(2020, 1, 1, 0, 59, 0, 0, time.UTC),
					To:   time.Date(2020, 1, 2, 0, 7, 0, 0, time.UTC),
				},
			},
			expectedSQL:   "SELECT * FROM table WHERE '2020-01-01T00:00:00Z'",
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, err := Interpolate(tt.configStruct)
			if tt.expectedError == "" {
				require.NoError(t, err)
				require.Equal(t, tt.expectedSQL, sql)
			} else {
				require.Error(t, err)
				require.Equal(t, tt.expectedError, err.Error())
			}
		})
	}
}
