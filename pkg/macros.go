package main

import (
	"fmt"
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
	timeRange := configStruct.TimeRange

	switch name {
	case "__time":
		if len(args) == 0 || args[0] == "" {
			return "", fmt.Errorf(missingColumnMessage, name)
		}
		return fmt.Sprintf("TRY_TO_TIMESTAMP_NTZ(%s) AS time", args[0]), nil
	case "__timeEpoch":
		if len(args) == 0 || args[0] == "" {
			return "", fmt.Errorf(missingColumnMessage, name)
		}
		return fmt.Sprintf("extract(epoch from %s) as time", args[0]), nil
	case "__timeFilter":
		if len(args) == 0 || args[0] == "" {
			return "", fmt.Errorf(missingColumnMessage, name)
		}
		column := args[0]
		timezone := "'UTC'"
		if len(args) > 1 {
			timezone = args[1]
		}
		return fmt.Sprintf("%s > CONVERT_TIMEZONE('UTC', %s, '%s'::timestamp_ntz) AND %s < CONVERT_TIMEZONE('UTC', %s, '%s'::timestamp_ntz)", column, timezone, timeRange.From.UTC().Format(time.RFC3339Nano), column, timezone, timeRange.To.UTC().Format(time.RFC3339Nano)), nil
	case "__timeTzFilter":
		if len(args) == 0 || args[0] == "" {
			return "", fmt.Errorf(missingColumnMessage, name)
		}
		column := args[0]
		return fmt.Sprintf("%s > '%s'::timestamp_tz AND %s < '%s'::timestamp_tz", column, timeRange.From.UTC().Format(time.RFC3339Nano), column, timeRange.To.UTC().Format(time.RFC3339Nano)), nil
	case "__timeFrom":
		return fmt.Sprintf("'%s'", timeRange.From.UTC().Format(time.RFC3339Nano)), nil
	case "__timeTo":
		return fmt.Sprintf("'%s'", timeRange.To.UTC().Format(time.RFC3339Nano)), nil
	case "__timeRoundFrom":
		//Rounds timestamp to the last 15min by default. First Argument could be passed to have a variable rounding in Minutes.
		timeSpan := 15
		if len(args) == 1 && args[0] != "" {
			if _, err := strconv.Atoi(args[0]); err == nil {
				timeSpan, _ = strconv.Atoi(args[0])
			} else {
				return "", fmt.Errorf("macro %v first argument must be a integer", name)
			}
			if timeSpan <= 0 {
				return "", fmt.Errorf("macro %v first argument must be a positive Integer", name)
			}
		} else if len(args) > 1 {
			return "", fmt.Errorf("macro %v only 1 argument allowed", name)
		}
		return fmt.Sprintf("'%s'", timeRange.From.UTC().Truncate(time.Minute*time.Duration(timeSpan)).Format(time.RFC3339Nano)), nil
	case "__timeRoundTo":
		timeSpan := 15
		if len(args) == 1 && args[0] != "" {
			if _, err := strconv.Atoi(args[0]); err == nil {
				timeSpan, _ = strconv.Atoi(args[0])
			} else {
				return "", fmt.Errorf("macro %v first argument must be a integer", name)
			}
			if timeSpan <= 0 {
				return "", fmt.Errorf("macro %v first argument must be a positive Integer", name)
			}
		} else if len(args) > 1 {
			return "", fmt.Errorf("macro %v only 1 argument allowed", name)
		}
		return fmt.Sprintf("'%s'", timeRange.To.UTC().Add(time.Minute*time.Duration(timeSpan)).Truncate(time.Minute*time.Duration(timeSpan)).Format(time.RFC3339Nano)), nil
	case "__timeGroup":
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

		return fmt.Sprintf("TIME_SLICE(TO_TIMESTAMP_NTZ(%s), %v, 'SECOND', 'START')", args[0], math.Max(1, interval.Seconds())), nil
	case "__timeGroupAlias":
		tg, err := evaluateMacro("__timeGroup", args, configStruct)
		if err == nil {
			return tg + " AS time", nil
		}
		return "", err
	case "__unixEpochFilter":
		if len(args) == 0 || args[0] == "" {
			return "", fmt.Errorf(missingColumnMessage, name)
		}
		return fmt.Sprintf("%s >= %d AND %s <= %d", args[0], timeRange.From.UTC().Unix(), args[0], timeRange.To.UTC().Unix()), nil
	case "__unixEpochNanoFilter":
		if len(args) == 0 || args[0] == "" {
			return "", fmt.Errorf(missingColumnMessage, name)
		}
		return fmt.Sprintf("%s >= %d AND %s <= %d", args[0], timeRange.From.UTC().UnixNano(), args[0], timeRange.To.UTC().UnixNano()), nil
	case "__unixEpochNanoFrom":
		return fmt.Sprintf("%d", timeRange.From.UTC().UnixNano()), nil
	case "__unixEpochNanoTo":
		return fmt.Sprintf("%d", timeRange.To.UTC().UnixNano()), nil
	case "__unixEpochGroup":
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
	case "__unixEpochGroupAlias":
		tg, err := evaluateMacro("__unixEpochGroup", args, configStruct)
		if err == nil {
			return tg + " AS time", nil
		}
		return "", err
	default:
		return "", fmt.Errorf("unknown macro %q", name)
	}
}
