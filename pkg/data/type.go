package data

import (
	"database/sql"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"time"
)

type QueryResult struct {
	Tables []Table
}

// DataTable structure containing columns and rows
type Table struct {
	Columns []*sql.ColumnType
	Rows    [][]interface{}
}

const TimeSeriesType = "time series"

func (qc *QueryConfigStruct) IsTimeSeriesType() bool {
	return qc.QueryType == TimeSeriesType
}

type QueryConfigStruct struct {
	FinalQuery    string
	QueryType     string
	RawQuery      string
	TimeColumns   []string
	TimeRange     backend.TimeRange
	Interval      time.Duration
	MaxDataPoints int64
	FillMode      string
	FillValue     float64
}
