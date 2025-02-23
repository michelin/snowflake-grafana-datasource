package query

import (
	"fmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/michelin/snowflake-grafana-datasource/pkg/data"
	"github.com/michelin/snowflake-grafana-datasource/pkg/utils"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const rsIdentifier = `([_a-zA-Z0-9]+)`
const sExpr = `\$` + rsIdentifier + `\(([^\)]*)\)`
const missingColumnMessage = "missing time column argument for macro %v"

func ReplaceAllStringSubmatchFunc(re *regexp.Regexp, str string, repl func([]string) string) string {
	result := ""
	lastIndex := 0

	for _, v := range re.FindAllSubmatchIndex([]byte(str), -1) {
		groups := []string{}
		for i := 0; i < len(v); i += 2 {
			groups = append(groups, str[v[i]:v[i+1]])
		}

		result += str[lastIndex:v[0]] + repl(groups)
		lastIndex = v[1]
	}

	return result + str[lastIndex:]
}

func Interpolate(configStruct *data.QueryConfigStruct) (string, error) {
	rExp, _ := regexp.Compile(sExpr)
	var macroError error

	sql := ReplaceAllStringSubmatchFunc(rExp, configStruct.RawQuery, func(groups []string) string {
		// Don't try to interpolate Snowflake macros SYSTEM$xxxxx
		if strings.Contains(configStruct.RawQuery, "SYSTEM"+groups[0]) {
			return groups[0]
		}
		args := strings.Split(groups[2], ",")
		for i, arg := range args {
			args[i] = strings.Trim(arg, " ")
		}
		res, err := evaluateMacro(groups[1], args, configStruct)
		if err != nil && macroError == nil {
			macroError = err
			return "macro_error()"
		}
		return res
	})

	if macroError != nil {
		return "", macroError
	}

	return sql, nil
}

func SetupFillmode(configStruct *data.QueryConfigStruct, fillmode string) error {
	switch fillmode {
	case "NULL":
		configStruct.FillMode = NullFill
	case "previous":
		configStruct.FillMode = PreviousFill
	default:
		configStruct.FillMode = ValueFill
		value, err := strconv.ParseFloat(fillmode, 64)
		if err != nil {
			return fmt.Errorf("error parsing fill value %v", fillmode)
		}
		configStruct.FillValue = value
	}

	return nil
}

// evaluateMacro convert macro expression to sql expression
func evaluateMacro(name string, args []string, configStruct *data.QueryConfigStruct) (string, error) {
	switch name {
	case "__time":
		return handleTimeMacro(args, name)
	case "__timeEpoch":
		return handleTimeEpochMacro(args, name)
	case "__timeFilter":
		return handleTimeFilterMacro(args, configStruct.TimeRange, name)
	case "__timeTzFilter":
		return handleTimeTzFilterMacro(args, configStruct.TimeRange, name)
	case "__timeFrom":
		return handleTimeFromMacro(configStruct.TimeRange)
	case "__timeTo":
		return handleTimeToMacro(configStruct.TimeRange)
	case "__timeRoundFrom":
		return handleTimeRoundFromMacro(args, configStruct.TimeRange, name)
	case "__timeRoundTo":
		return handleTimeRoundToMacro(args, configStruct.TimeRange, name)
	case "__timeGroup":
		return handleTimeGroupMacro(args, configStruct, name)
	case "__timeGroupAlias":
		return handleTimeGroupAliasMacro(args, configStruct)
	case "__unixEpochFilter":
		return handleUnixEpochFilterMacro(args, configStruct.TimeRange, name)
	case "__unixEpochNanoFilter":
		return handleUnixEpochNanoFilterMacro(args, configStruct.TimeRange, name)
	case "__unixEpochNanoFrom":
		return handleUnixEpochNanoFromMacro(configStruct.TimeRange)
	case "__unixEpochNanoTo":
		return handleUnixEpochNanoToMacro(configStruct.TimeRange)
	case "__unixEpochGroup":
		return handleUnixEpochGroupMacro(args, configStruct, name)
	case "__unixEpochGroupAlias":
		return handleUnixEpochGroupAliasMacro(args, configStruct)
	default:
		return "", fmt.Errorf("unknown macro %q", name)
	}
}

func handleTimeMacro(args []string, name string) (string, error) {
	if len(args) == 0 || args[0] == "" {
		return "", fmt.Errorf(missingColumnMessage, name)
	}
	return fmt.Sprintf("TRY_TO_TIMESTAMP_NTZ(%s) AS time", args[0]), nil
}

func handleTimeEpochMacro(args []string, name string) (string, error) {
	if len(args) == 0 || args[0] == "" {
		return "", fmt.Errorf(missingColumnMessage, name)
	}
	return fmt.Sprintf("extract(epoch from %s) as time", args[0]), nil
}

func handleTimeFilterMacro(args []string, timeRange backend.TimeRange, name string) (string, error) {
	if len(args) == 0 || args[0] == "" {
		return "", fmt.Errorf(missingColumnMessage, name)
	}
	column := args[0]
	timezone := "'UTC'"
	if len(args) > 1 {
		timezone = args[1]
	}
	return fmt.Sprintf("%s > CONVERT_TIMEZONE('UTC', %s, '%s'::timestamp_ntz) AND %s < CONVERT_TIMEZONE('UTC', %s, '%s'::timestamp_ntz)", column, timezone, timeRange.From.UTC().Format(time.RFC3339Nano), column, timezone, timeRange.To.UTC().Format(time.RFC3339Nano)), nil
}

func handleTimeTzFilterMacro(args []string, timeRange backend.TimeRange, name string) (string, error) {
	if len(args) == 0 || args[0] == "" {
		return "", fmt.Errorf(missingColumnMessage, name)
	}
	column := args[0]
	return fmt.Sprintf("%s > '%s'::timestamp_tz AND %s < '%s'::timestamp_tz", column, timeRange.From.UTC().Format(time.RFC3339Nano), column, timeRange.To.UTC().Format(time.RFC3339Nano)), nil
}

func handleTimeFromMacro(timeRange backend.TimeRange) (string, error) {
	return fmt.Sprintf("'%s'", timeRange.From.UTC().Format(time.RFC3339Nano)), nil
}

func handleTimeToMacro(timeRange backend.TimeRange) (string, error) {
	return fmt.Sprintf("'%s'", timeRange.To.UTC().Format(time.RFC3339Nano)), nil
}

func handleTimeRoundFromMacro(args []string, timeRange backend.TimeRange, name string) (string, error) {
	timeSpan, err := parseTimeSpan(args, name)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("'%s'", timeRange.From.UTC().Truncate(time.Minute*time.Duration(timeSpan)).Format(time.RFC3339Nano)), nil
}

func handleTimeRoundToMacro(args []string, timeRange backend.TimeRange, name string) (string, error) {
	timeSpan, err := parseTimeSpan(args, name)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("'%s'", timeRange.To.UTC().Add(time.Minute*time.Duration(timeSpan)).Truncate(time.Minute*time.Duration(timeSpan)).Format(time.RFC3339Nano)), nil
}

func handleTimeGroupMacro(args []string, configStruct *data.QueryConfigStruct, name string) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("macro %v needs time column and interval and optional fill value", name)
	}
	interval, err := utils.ParseInterval(strings.Trim(args[1], `'`))
	if err != nil {
		return "", fmt.Errorf("error parsing interval %v", args[1])
	}
	if len(args) > 2 {
		err := SetupFillmode(configStruct, args[2])
		if err != nil {
			return "", err
		}
	}

	timeExpr := fmt.Sprintf("TO_TIMESTAMP_NTZ(%s)", args[0])
    if len(args) > 3 {
        timeExpr = fmt.Sprintf("TO_TIMESTAMP_NTZ(CONVERT_TIMEZONE(%s, %s))", args[3], args[0])
    }

	duration := interval.Seconds()
	timeUnit := "SECOND"

	// If the interval can be translated to weeks exactly, then use WEEK as time slice unit as it allows users to configure which day they want the graphs to be based on 
	// as opposed to having them always start on Thursdays due to 1970-01-01 being thursday.
	const WEEK_IN_SECONDS = 7 * 24 * 3600
	if interval.Seconds() > 1 && int64(interval.Seconds()) % WEEK_IN_SECONDS == 0 {
		duration = interval.Seconds() / WEEK_IN_SECONDS
		timeUnit = "WEEK"
	}

	return fmt.Sprintf("TIME_SLICE(%s, %v, '%s', 'START')", timeExpr, math.Max(1, duration), timeUnit), nil
}

func handleTimeGroupAliasMacro(args []string, configStruct *data.QueryConfigStruct) (string, error) {
	tg, err := handleTimeGroupMacro(args, configStruct, "__timeGroup")
	if err == nil {
		return tg + " AS time", nil
	}
	return "", err
}

func handleUnixEpochFilterMacro(args []string, timeRange backend.TimeRange, name string) (string, error) {
	if len(args) == 0 || args[0] == "" {
		return "", fmt.Errorf(missingColumnMessage, name)
	}
	return fmt.Sprintf("%s >= %d AND %s <= %d", args[0], timeRange.From.UTC().Unix(), args[0], timeRange.To.UTC().Unix()), nil
}

func handleUnixEpochNanoFilterMacro(args []string, timeRange backend.TimeRange, name string) (string, error) {
	if len(args) == 0 || args[0] == "" {
		return "", fmt.Errorf(missingColumnMessage, name)
	}
	return fmt.Sprintf("%s >= %d AND %s <= %d", args[0], timeRange.From.UTC().UnixNano(), args[0], timeRange.To.UTC().UnixNano()), nil
}

func handleUnixEpochNanoFromMacro(timeRange backend.TimeRange) (string, error) {
	return fmt.Sprintf("%d", timeRange.From.UTC().UnixNano()), nil
}

func handleUnixEpochNanoToMacro(timeRange backend.TimeRange) (string, error) {
	return fmt.Sprintf("%d", timeRange.To.UTC().UnixNano()), nil
}

func handleUnixEpochGroupMacro(args []string, configStruct *data.QueryConfigStruct, name string) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("macro %v needs time column and interval and optional fill value", name)
	}
	interval, err := utils.ParseInterval(strings.Trim(args[1], `'`))
	if err != nil {
		return "", fmt.Errorf("error parsing interval %v", args[1])
	}
	if len(args) == 3 {
		err := SetupFillmode(configStruct, args[2])
		if err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("floor(%s/%v)*%v", args[0], interval.Seconds(), interval.Seconds()), nil
}

func handleUnixEpochGroupAliasMacro(args []string, configStruct *data.QueryConfigStruct) (string, error) {
	tg, err := handleUnixEpochGroupMacro(args, configStruct, "__unixEpochGroup")
	if err == nil {
		return tg + " AS time", nil
	}
	return "", err
}

func parseTimeSpan(args []string, name string) (int, error) {
	timeSpan := 15
	if len(args) == 1 && args[0] != "" {
		if _, err := strconv.Atoi(args[0]); err == nil {
			timeSpan, _ = strconv.Atoi(args[0])
		} else {
			return 0, fmt.Errorf("macro %v first argument must be a integer", name)
		}
		if timeSpan <= 0 {
			return 0, fmt.Errorf("macro %v first argument must be a positive Integer", name)
		}
	} else if len(args) > 1 {
		return 0, fmt.Errorf("macro %v only 1 argument allowed", name)
	}
	return timeSpan, nil
}
