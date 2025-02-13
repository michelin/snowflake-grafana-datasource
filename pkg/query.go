package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	_data "github.com/michelin/snowflake-grafana-datasource/pkg/data"
	"github.com/michelin/snowflake-grafana-datasource/pkg/query"
	"github.com/michelin/snowflake-grafana-datasource/pkg/utils"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

const rowLimit = 1000000

// type
var boolean bool
var tim time.Time
var float float64
var str string
var integer int64

type queryModel struct {
	QueryText   string   `json:"queryText"`
	QueryType   string   `json:"queryType"`
	TimeColumns []string `json:"timeColumns"`
	FillMode    string   `json:"fillMode"`
}

func fetchData(ctx context.Context, qc *_data.QueryConfigStruct, config *pluginConfig, authenticationSecret _data.AuthenticationSecret) (result _data.QueryResult, err error) {
	connectionString := getConnectionString(config, authenticationSecret)

	db, err := sql.Open("snowflake", connectionString)
	if err != nil {
		log.DefaultLogger.Error("Could not open database", "err", err)
		return result, err
	}
	defer db.Close()

	log.DefaultLogger.Info("Query", "finalQuery", qc.FinalQuery)
	rows, err := db.QueryContext(utils.AddQueryTagInfos(ctx, qc), qc.FinalQuery)
	if err != nil {
		if strings.Contains(err.Error(), "000605") {
			log.DefaultLogger.Info("Query got cancelled", "query", qc.FinalQuery, "err", err)
			return result, err
		}

		log.DefaultLogger.Error("Could not execute query", "query", qc.FinalQuery, "err", err)
		return result, err
	}
	defer rows.Close()

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
		values, err := transformQueryResult(*qc, columnTypes, rows)
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

func transformQueryResult(qc _data.QueryConfigStruct, columnTypes []*sql.ColumnType, rows *sql.Rows) ([]interface{}, error) {
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
		if qc.IsTimeSeriesType() && utils.EqualsIgnoreCase(qc.TimeColumns, columnTypes[i].Name()) && reflect.TypeOf(values[i]) == reflect.TypeOf(str) {
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

func (td *SnowflakeDatasource) query(ctx context.Context, dataQuery backend.DataQuery, request *backend.QueryDataRequest, config pluginConfig, authentication _data.AuthenticationSecret) (response backend.DataResponse) {
	var qm queryModel
	err := json.Unmarshal(dataQuery.JSON, &qm)
	if err != nil {
		log.DefaultLogger.Error("Could not unmarshal query", "err", err)
		response.Error = err
		return response
	}

	if qm.QueryText == "" {
		log.DefaultLogger.Error("SQL query must no be empty")
		response.Error = fmt.Errorf("SQL query must no be empty")
		return response
	}

	queryConfig := _data.QueryConfigStruct{
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
	}

	log.DefaultLogger.Info("Query config", "config", qm)

	// Apply macros
	queryConfig.FinalQuery, err = query.Interpolate(&queryConfig)
	if err != nil {
		response.Error = err
		return response
	}

	// Remove final semi column
	queryConfig.FinalQuery = strings.TrimSuffix(strings.TrimSpace(queryConfig.FinalQuery), ";")

	frame := data.NewFrame("")
	dataResponse, err := fetchData(ctx, &queryConfig, &config, authentication)
	if err != nil {
		response.Error = err
		return response
	}
	log.DefaultLogger.Debug("Response", "data", dataResponse)
	for _, table := range dataResponse.Tables {
		timeColumnIndex := -1
		for i, column := range table.Columns {
			// Check time column
			if queryConfig.IsTimeSeriesType() && utils.EqualsIgnoreCase(queryConfig.TimeColumns, column.Name()) {
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
				fillTimesSeries(queryConfig, intervalStart, row[utils.Max(int64(timeColumnIndex), 0)].(time.Time).UnixNano()/1e6, timeColumnIndex, frame, len(table.Columns), &count, utils.PreviousRow(table.Rows, j))
			}
			// without fill mode
			for i, v := range row {
				utils.InsertFrameField(frame, v, i)
			}
			count++
		}
		fillTimesSeries(queryConfig, intervalStart, intervalEnd, timeColumnIndex, frame, len(table.Columns), &count, utils.PreviousRow(table.Rows, len(table.Rows)))
	}
	if queryConfig.IsTimeSeriesType() {
		frame, err = td.longToWide(frame, queryConfig, dataResponse)
		if err != nil {
			response.Error = err
			return response
		}
	}
	log.DefaultLogger.Debug("Converted wide time Frame is:", frame)
	frame.RefID = dataQuery.RefID
	frame.Meta = &data.FrameMeta{
		Type:                data.FrameTypeTimeSeriesWide,
		ExecutedQueryString: queryConfig.FinalQuery,
	}

	response.Frames = append(response.Frames, frame)

	return response
}

func (td *SnowflakeDatasource) longToWide(frame *data.Frame, queryConfig _data.QueryConfigStruct, dataResponse _data.QueryResult) (*data.Frame, error) {
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
