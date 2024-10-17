package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	sf "github.com/snowflakedb/gosnowflake"
)

const rowLimit = 1000000

const timeSeriesType = "time series"

func (qc *queryConfigStruct) isTimeSeriesType() bool {
	return qc.QueryType == timeSeriesType
}

type queryCounter int32

func (c *queryCounter) inc() int32 {
	return atomic.AddInt32((*int32)(c), 1)
}

func (c *queryCounter) dec() int32 {
	return atomic.AddInt32((*int32)(c), -1)
}

func (c *queryCounter) get() int32 {
	return atomic.LoadInt32((*int32)(c))
}

type queryConfigStruct struct {
	FinalQuery    string
	QueryType     string
	RawQuery      string
	TimeColumns   []string
	TimeRange     backend.TimeRange
	Interval      time.Duration
	MaxDataPoints int64
	FillMode      string
	FillValue     float64
	db            *sql.DB
	config        *pluginConfig
	actQueryCount *queryCounter
}

// type
var boolean bool
var tim time.Time
var float float64
var str string
var integer int64

// Constant used to describe the time series fill mode if no value has been seen
const (
	NullFill     = "null"
	PreviousFill = "previous"
	ValueFill    = "value"
)

type queryModel struct {
	QueryText   string   `json:"queryText"`
	QueryType   string   `json:"queryType"`
	TimeColumns []string `json:"timeColumns"`
	FillMode    string   `json:"fillMode"`
}

func (qc *queryConfigStruct) fetchData(ctx context.Context) (result DataQueryResult, err error) {
	qc.actQueryCount.inc()
	// Custom configuration to reduce memory footprint
	sf.MaxChunkDownloadWorkers = 2
	sf.CustomJSONDecoderEnabled = true

	start := time.Now()
	stats := qc.db.Stats()
	defer func() {
		qc.actQueryCount.dec()
		duration := time.Since(start)
		log.DefaultLogger.Info(fmt.Sprintf("%+v - %s - %d", stats, duration, int(qc.actQueryCount.get())))

	}()
	if int(qc.config.IntMaxQueuedQueries) > 0 && int(qc.actQueryCount.get()) >= (int(qc.config.IntMaxQueuedQueries)) {
		err := errors.New("too many queries in queue. Check Snowflake connectivity or increase MaxQueuedQeries count")
		log.DefaultLogger.Error("Poolsize exceeded", "query", qc.FinalQuery, "err", err)
		return result, err
	}
	rows, err := qc.db.QueryContext(ctx, qc.FinalQuery)
	if err != nil {
		if strings.Contains(err.Error(), "000605") {
			log.DefaultLogger.Info("Query got cancelled", "query", qc.FinalQuery, "err", err)
			return result, err
		}

		log.DefaultLogger.Error("Could not execute query", "query", qc.FinalQuery, "err", err)
		return result, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.DefaultLogger.Warn("Failed to close rows", "err", err)
		}
	}()

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		log.DefaultLogger.Error("Could not get column types", "err", err)
		return result, err
	}
	columnCount := len(columnTypes)

	if columnCount == 0 {
		return result, nil
	}

	table := DataTable{
		Columns: columnTypes,
		Rows:    make([][]interface{}, 0),
	}

	rowCount := 0
	for ; rows.Next(); rowCount++ {
		if rowCount > rowLimit {
			return result, fmt.Errorf("query row limit exceeded, limit %d", rowLimit)
		}
		values, err := qc.transformQueryResult(columnTypes, rows)
		if err != nil {
			return result, err
		}
		table.Rows = append(table.Rows, values)
	}

	err = rows.Err()
	if err != nil {
		log.DefaultLogger.Error("The row scan finished with an error", "err", err)
		return result, err
	}

	result.Tables = append(result.Tables, table)
	return result, nil
}

func (qc *queryConfigStruct) transformQueryResult(columnTypes []*sql.ColumnType, rows *sql.Rows) ([]interface{}, error) {
	values := make([]interface{}, len(columnTypes))
	valuePtrs := make([]interface{}, len(columnTypes))

	for i := 0; i < len(columnTypes); i++ {
		valuePtrs[i] = &values[i]
	}

	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, err
	}

	column_types, _ := rows.ColumnTypes()

	// convert types from string type to real type
	for i := 0; i < len(columnTypes); i++ {
		log.DefaultLogger.Debug("Type", fmt.Sprintf("%T %v ", values[i], values[i]), columnTypes[i].DatabaseTypeName())

		// Convert time columns when query mode is time series
		if qc.isTimeSeriesType() && equalsIgnoreCase(qc.TimeColumns, columnTypes[i].Name()) && reflect.TypeOf(values[i]) == reflect.TypeOf(str) {
			if v, err := strconv.ParseFloat(values[i].(string), 64); err == nil {
				values[i] = time.Unix(int64(v), 0)
			} else {
				return nil, fmt.Errorf("column %s cannot be converted to Time", columnTypes[i].Name())
			}
			continue
		}

		if values[i] != nil {
			switch column_types[i].ScanType() {
			case reflect.TypeOf(boolean):
				values[i] = values[i].(bool)
			case reflect.TypeOf(tim):
				values[i] = values[i].(time.Time)
			case reflect.TypeOf(integer):
				n := new(big.Float)
				n.SetString(values[i].(string))
				precision, _, _ := columnTypes[i].DecimalSize()
				if precision > 1 {
					values[i], _ = n.Float64()
				} else {
					values[i], _ = n.Int64()
				}
			case reflect.TypeOf(float):
				if reflect.TypeOf(float) == reflect.TypeOf(values[i]) {
					values[i] = values[i].(float64)
				} else if v, err := strconv.ParseFloat(values[i].(string), 64); err == nil {
					values[i] = v
				} else {
					log.DefaultLogger.Info("Rows", "Error converting string to float64", values[i])
				}
			case reflect.TypeOf(str):
				values[i] = values[i].(string)
			default:
				values[i] = values[i].(string)
			}
		}
	}

	return values, nil
}

func (td *SnowflakeDatasource) query(ctx context.Context, wg *sync.WaitGroup, ch chan DBDataResponse, instance *instanceSettings, dataQuery backend.DataQuery) {
	defer wg.Done()
	queryResult := DBDataResponse{
		dataResponse: backend.DataResponse{},
		refID:        dataQuery.RefID,
	}

	defer func() {
		if r := recover(); r != nil {
			log.DefaultLogger.Error("ExecuteQuery panic", "error", r, "stack", string(debug.Stack()))
			if theErr, ok := r.(error); ok {
				queryResult.dataResponse.Error = theErr
			} else if theErrString, ok := r.(string); ok {
				queryResult.dataResponse.Error = fmt.Errorf(theErrString)
			}
			ch <- queryResult
		}
	}()

	var qm queryModel
	err := json.Unmarshal(dataQuery.JSON, &qm)
	if err != nil {
		panic("Could not unmarshal query")
	}

	if qm.QueryText == "" {
		panic("Query model property rawSql should not be empty at this point")
	}

	queryConfig := queryConfigStruct{
		FinalQuery:    qm.QueryText,
		RawQuery:      qm.QueryText,
		TimeColumns:   qm.TimeColumns,
		FillMode:      qm.FillMode,
		QueryType:     dataQuery.QueryType,
		Interval:      dataQuery.Interval,
		TimeRange:     dataQuery.TimeRange,
		MaxDataPoints: dataQuery.MaxDataPoints,
		db:            instance.db,
		config:        instance.config,
		actQueryCount: &instance.actQueryCount,
	}

	errAppendDebug := func(frameErr string, err error, query string) {
		var emptyFrame data.Frame
		emptyFrame.SetMeta(&data.FrameMeta{
			ExecutedQueryString: query,
		})
		queryResult.dataResponse.Error = fmt.Errorf("%s: %w", frameErr, err)
		queryResult.dataResponse.Frames = data.Frames{&emptyFrame}
		ch <- queryResult
	}

	// Apply macros
	queryConfig.FinalQuery, err = Interpolate(&queryConfig)
	if err != nil {
		errAppendDebug("interpolation failed", err, queryConfig.FinalQuery)
		return
	}

	// Remove final semi column
	queryConfig.FinalQuery = strings.TrimSuffix(strings.TrimSpace(queryConfig.FinalQuery), ";")

	frame := data.NewFrame("")
	dataResponse, err := queryConfig.fetchData(ctx)
	if err != nil {
		errAppendDebug("db query error", err, queryConfig.FinalQuery)
		return
	}
	log.DefaultLogger.Debug("Response", "data", dataResponse)
	for _, table := range dataResponse.Tables {
		timeColumnIndex := -1
		for i, column := range table.Columns {
			if err != nil {
				errAppendDebug("db query error", err, queryConfig.FinalQuery)
				return
			}
			// Check time column
			if queryConfig.isTimeSeriesType() && equalsIgnoreCase(queryConfig.TimeColumns, column.Name()) {
				if strings.EqualFold(column.Name(), "Time") {
					timeColumnIndex = i
				}
				frame.Fields = append(frame.Fields, data.NewField(column.Name(), nil, []*time.Time{}))
				continue
			}
			switch column.ScanType() {
			case reflect.TypeOf(boolean):
				frame.Fields = append(frame.Fields, data.NewField(column.Name(), nil, []*bool{}))
			case reflect.TypeOf(tim):
				frame.Fields = append(frame.Fields, data.NewField(column.Name(), nil, []*time.Time{}))
			case reflect.TypeOf(integer):
				precision, _, _ := column.DecimalSize()
				if precision > 1 {
					frame.Fields = append(frame.Fields, data.NewField(column.Name(), nil, []*float64{}))
				} else {
					frame.Fields = append(frame.Fields, data.NewField(column.Name(), nil, []*int64{}))
				}
			case reflect.TypeOf(float):
				frame.Fields = append(frame.Fields, data.NewField(column.Name(), nil, []*float64{}))
			case reflect.TypeOf(str):
				frame.Fields = append(frame.Fields, data.NewField(column.Name(), nil, []*string{}))
			default:
				log.DefaultLogger.Error("Rows", "Unknown database type", column.DatabaseTypeName())
				frame.Fields = append(frame.Fields, data.NewField(column.Name(), nil, []*string{}))
			}
		}

		intervalStart := queryConfig.TimeRange.From.UnixNano() / 1e6
		intervalEnd := queryConfig.TimeRange.To.UnixNano() / 1e6

		count := 0
		// add rows
		for j, row := range table.Rows {
			// Handle fill mode when the time column exist
			if timeColumnIndex != -1 {
				fillTimesSeries(queryConfig, intervalStart, row[Max(int64(timeColumnIndex), 0)].(time.Time).UnixNano()/1e6, timeColumnIndex, frame, len(table.Columns), &count, previousRow(table.Rows, j))
			}
			// without fill mode
			for i, v := range row {
				insertFrameField(frame, v, i)
			}
			count++
		}
		fillTimesSeries(queryConfig, intervalStart, intervalEnd, timeColumnIndex, frame, len(table.Columns), &count, previousRow(table.Rows, len(table.Rows)))
	}
	if queryConfig.isTimeSeriesType() {
		frame, err = td.longToWide(frame, queryConfig, dataResponse)
		if err != nil {
			queryResult.dataResponse.Error = fmt.Errorf("%w", err)
			queryResult.dataResponse.Frames = data.Frames{frame}
			ch <- queryResult
		}
	}
	log.DefaultLogger.Debug("Converted wide time Frame is:", frame)
	frame.RefID = dataQuery.RefID
	frame.Meta = &data.FrameMeta{
		Type:                data.FrameTypeTimeSeriesWide,
		ExecutedQueryString: queryConfig.FinalQuery,
	}

	queryResult.dataResponse.Frames = data.Frames{frame}
	ch <- queryResult
}

func (td *SnowflakeDatasource) longToWide(frame *data.Frame, queryConfig queryConfigStruct, dataResponse DataQueryResult) (*data.Frame, error) {
	tsSchema := frame.TimeSeriesSchema()
	if tsSchema.Type == data.TimeSeriesTypeLong {
		fillMode := &data.FillMissing{Mode: mapFillMode(queryConfig.FillMode), Value: queryConfig.FillValue}
		if len(dataResponse.Tables) > 0 && len(dataResponse.Tables[0].Rows) > 0 {
			var err error
			frame, err = data.LongToWide(frame, fillMode)
			if err != nil {
				log.DefaultLogger.Error("Could not convert long frame to wide frame", "err", err)
				return nil, err
			}
		}
		for _, field := range frame.Fields {
			if field.Labels != nil {
				for _, val := range field.Labels {
					field.Name += "_" + string(val)
				}
			}
		}
	}
	return frame, nil
}

func mapFillMode(fillModeString string) data.FillMode {
	var fillMode = data.FillModeNull
	switch fillModeString {
	case ValueFill:
		fillMode = data.FillModeValue
	case NullFill:
		fillMode = data.FillModeNull
	case PreviousFill:
		fillMode = data.FillModePrevious
	default:
		// no-op
	}
	return fillMode
}

func fillTimesSeries(queryConfig queryConfigStruct, intervalStart int64, intervalEnd int64, timeColumnIndex int, frame *data.Frame, columnSize int, count *int, previousRow []interface{}) {
	if queryConfig.isTimeSeriesType() && queryConfig.FillMode != "" && timeColumnIndex != -1 {
		for stepTime := intervalStart + queryConfig.Interval.Nanoseconds()/1e6*int64(*count); stepTime < intervalEnd; stepTime = stepTime + (queryConfig.Interval.Nanoseconds() / 1e6) {
			for i := 0; i < columnSize; i++ {
				if i == timeColumnIndex {
					t := time.Unix(stepTime/1e3, 0)
					frame.Fields[i].Append(&t)
					continue
				}
				switch queryConfig.FillMode {
				case ValueFill:
					frame.Fields[i].Append(&queryConfig.FillValue)
				case NullFill:
					frame.Fields[i].Append(nil)
				case PreviousFill:
					if previousRow == nil {
						insertFrameField(frame, nil, i)
					} else {
						insertFrameField(frame, previousRow[i], i)
					}
				default:
				}
			}
			*count++
		}
	}
}
