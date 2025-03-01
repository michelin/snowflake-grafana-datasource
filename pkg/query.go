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
	_data "github.com/michelin/snowflake-grafana-datasource/pkg/data"
	"github.com/michelin/snowflake-grafana-datasource/pkg/query"
	"github.com/michelin/snowflake-grafana-datasource/pkg/utils"
)

const rowLimit = 1000000

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

type QueryStruct struct {
	qc                  *_data.QueryConfigStruct
	db                  *sql.DB
	intMaxQueuedQueries int64
	queryCounter        queryCounter
}

type queryModel struct {
	QueryText   string   `json:"queryText"`
	QueryType   string   `json:"queryType"`
	TimeColumns []string `json:"timeColumns"`
	FillMode    string   `json:"fillMode"`
}

// type
var boolean bool
var tim time.Time
var float float64
var str string
var integer int64

func (td *SnowflakeDatasource) query(ctx context.Context, wg *sync.WaitGroup, ch chan DBDataResponse, request *backend.QueryDataRequest, instance *instanceSettings, dataQuery backend.DataQuery) {
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

	queryStruct := QueryStruct{
		qc: &_data.QueryConfigStruct{
			FinalQuery:    qm.QueryText,
			RawQuery:      qm.QueryText,
			TimeColumns:   qm.TimeColumns,
			FillMode:      qm.FillMode,
			QueryType:     dataQuery.QueryType,
			Interval:      dataQuery.Interval,
			TimeRange:     dataQuery.TimeRange,
			MaxDataPoints: dataQuery.MaxDataPoints,
			DashboardId:   request.GetHTTPHeader("X-Dashboard-Uid"),
			PanelId:       request.GetHTTPHeader("X-Panel-Id"),
		},
		intMaxQueuedQueries: instance.config.IntMaxQueuedQueries,
		db:                  instance.db,
		queryCounter:        instance.actQueryCount}

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
	queryStruct.qc.FinalQuery, err = query.Interpolate(queryStruct.qc)
	if err != nil {
		errAppendDebug("interpolation failed", err, queryStruct.qc.FinalQuery)
		return
	}

	// Remove final semi column
	queryStruct.qc.FinalQuery = strings.TrimSuffix(strings.TrimSpace(queryStruct.qc.FinalQuery), ";")

	frame := data.NewFrame("")
	dataResponse, err := queryStruct.fetchData(ctx)
	if err != nil {
		errAppendDebug("db query error", err, queryStruct.qc.FinalQuery)
		return
	}
	log.DefaultLogger.Debug("Response", "data", dataResponse)
	for _, table := range dataResponse.Tables {
		timeColumnIndex := -1
		for i, column := range table.Columns {
			if err != nil {
				errAppendDebug("db query error", err, queryStruct.qc.FinalQuery)
				return
			}
			// Check time column
			if queryStruct.qc.IsTimeSeriesType() && utils.EqualsIgnoreCase(queryStruct.qc.TimeColumns, column.Name()) {
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

		intervalStart := queryStruct.qc.TimeRange.From.UnixNano() / 1e6
		intervalEnd := queryStruct.qc.TimeRange.To.UnixNano() / 1e6

		count := 0
		// add rows
		for j, row := range table.Rows {
			// Handle fill mode when the time column exist
			if timeColumnIndex != -1 {
				fillTimesSeries(*queryStruct.qc, intervalStart, row[utils.Max(int64(timeColumnIndex), 0)].(time.Time).UnixNano()/1e6, timeColumnIndex, frame, len(table.Columns), &count, utils.PreviousRow(table.Rows, j))
			}
			// without fill mode
			for i, v := range row {
				utils.InsertFrameField(frame, v, i)
			}
			count++
		}
		fillTimesSeries(*queryStruct.qc, intervalStart, intervalEnd, timeColumnIndex, frame, len(table.Columns), &count, utils.PreviousRow(table.Rows, len(table.Rows)))
	}
	if queryStruct.qc.IsTimeSeriesType() {
		frame, err = longToWide(frame, *queryStruct.qc, dataResponse)
		if err != nil {
			errAppendDebug("db transformation error", err, queryStruct.qc.FinalQuery)
			return
		}
	}
	log.DefaultLogger.Debug("Converted wide time Frame is:", frame)
	frame.RefID = dataQuery.RefID
	frame.Meta = &data.FrameMeta{
		Type:                data.FrameTypeTimeSeriesWide,
		ExecutedQueryString: queryStruct.qc.FinalQuery,
	}

	queryResult.dataResponse.Frames = data.Frames{frame}
	ch <- queryResult
}

func (qs *QueryStruct) fetchData(ctx context.Context) (result _data.QueryResult, err error) {
	qs.queryCounter.inc()

	start := time.Now()
	stats := qs.db.Stats()
	defer func() {
		if qs.queryCounter.get() > 0 {
			qs.queryCounter.dec()
		}

		duration := time.Since(start)
		log.DefaultLogger.Debug(fmt.Sprintf("%+v - %s - %d", stats, duration, int(qs.queryCounter.get())))

	}()
	if int(qs.intMaxQueuedQueries) > 0 && int(qs.queryCounter.get()) >= (int(qs.intMaxQueuedQueries)) {
		err := errors.New("too many queries in queue. Check Snowflake connectivity or increase MaxQueuedQeries count")
		log.DefaultLogger.Error("Poolsize exceeded", "query", qs.qc.FinalQuery, "err", err)
		return result, err
	}
	rows, err := qs.db.QueryContext(utils.AddQueryTagInfos(ctx, qs.qc), qs.qc.FinalQuery)
	if err != nil {
		if strings.Contains(err.Error(), "000605") {
			log.DefaultLogger.Info("Query got cancelled", "query", qs.qc.FinalQuery, "err", err)
			return result, err
		}

		log.DefaultLogger.Error("Could not execute query", "query", qs.qc.FinalQuery, "err", err)
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

	table := _data.Table{
		Columns: columnTypes,
		Rows:    make([][]interface{}, 0),
	}

	rowCount := 0
	for ; rows.Next(); rowCount++ {
		if rowCount > rowLimit {
			return result, fmt.Errorf("query row limit exceeded, limit %d", rowLimit)
		}
		values, err := qs.transformQueryResult(columnTypes, rows)
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

func (qs *QueryStruct) transformQueryResult(columnTypes []*sql.ColumnType, rows *sql.Rows) ([]interface{}, error) {
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
		if qs.qc.IsTimeSeriesType() && utils.EqualsIgnoreCase(qs.qc.TimeColumns, columnTypes[i].Name()) && reflect.TypeOf(values[i]) == reflect.TypeOf(str) {
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

func fillTimesSeries(queryConfig _data.QueryConfigStruct, intervalStart int64, intervalEnd int64, timeColumnIndex int, frame *data.Frame, columnSize int, count *int, previousRow []interface{}) {
	if queryConfig.IsTimeSeriesType() && queryConfig.FillMode != "" && timeColumnIndex != -1 {
		for stepTime := intervalStart + queryConfig.Interval.Nanoseconds()/1e6*int64(*count); stepTime < intervalEnd; stepTime = stepTime + (queryConfig.Interval.Nanoseconds() / 1e6) {
			for i := 0; i < columnSize; i++ {
				if i == timeColumnIndex {
					t := time.Unix(stepTime/1e3, 0)
					frame.Fields[i].Append(&t)
					continue
				}
				switch queryConfig.FillMode {
				case query.ValueFill:
					frame.Fields[i].Append(&queryConfig.FillValue)
				case query.NullFill:
					frame.Fields[i].Append(nil)
				case query.PreviousFill:
					if previousRow == nil {
						utils.InsertFrameField(frame, nil, i)
					} else {
						utils.InsertFrameField(frame, previousRow[i], i)
					}
				default:
				}
			}
			*count++
		}
	}
}

func longToWide(frame *data.Frame, queryConfig _data.QueryConfigStruct, dataResponse _data.QueryResult) (*data.Frame, error) {
	tsSchema := frame.TimeSeriesSchema()
	if tsSchema.Type == data.TimeSeriesTypeLong {
		fillMode := &data.FillMissing{Mode: query.MapFillMode(queryConfig.FillMode), Value: queryConfig.FillValue}
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
