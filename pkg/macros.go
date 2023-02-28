package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const rsIdentifier = `([_a-zA-Z0-9]+)`
const sExpr = `\$` + rsIdentifier + `\(([^\)]*)\)`

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

func Interpolate(configStruct *queryConfigStruct) (string, error) {
	rExp, _ := regexp.Compile(sExpr)
	var macroError error

	sql := ReplaceAllStringSubmatchFunc(rExp, configStruct.RawQuery, func(groups []string) string {
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

func SetupFillmode(configStruct *queryConfigStruct, fillmode string) error {
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
func evaluateMacro(name string, args []string, configStruct *queryConfigStruct) (string, error) {
	timeRange := configStruct.TimeRange
	switch name {
	case "__time":
		if len(args) == 0 {
			return "", fmt.Errorf("missing time column argument for macro %v", name)
		}
		return fmt.Sprintf("TRY_TO_TIMESTAMP(%s) AS time", args[0]), nil
	case "__timeEpoch":
		if len(args) == 0 {
			return "", fmt.Errorf("missing time column argument for macro %v", name)
		}
		return fmt.Sprintf("extract(epoch from %s) as time", args[0]), nil
	case "__timeFilter":
		if len(args) == 0 {
			return "", fmt.Errorf("missing time column argument for macro %v", name)
		}
		column := args[0]
		timezone := "'UTC'"
		if len(args) == 1 {
			return fmt.Sprintf("%s BETWEEN '%s' AND '%s'", args[0], timeRange.From.UTC().Format(time.RFC3339Nano), timeRange.To.UTC().Format(time.RFC3339Nano)), nil
		} else {
			timezone = args[1]
			return fmt.Sprintf("CONVERT_TIMEZONE('UTC', %s, %s) >= '%s' AND CONVERT_TIMEZONE('UTC', %s, %s) <= '%s'", timezone, column, timeRange.From.UTC().Format(time.RFC3339Nano), timezone, column, timeRange.To.UTC().Format(time.RFC3339Nano)), nil
		}

	case "__timeFrom":
		return fmt.Sprintf("'%s'", timeRange.From.UTC().Format(time.RFC3339Nano)), nil
	case "__timeTo":
		return fmt.Sprintf("'%s'", timeRange.To.UTC().Format(time.RFC3339Nano)), nil
	case "__timeGroup":
		if len(args) < 2 {
			return "", fmt.Errorf("macro %v needs time column and interval and optional fill value", name)
		}
		interval, err := ParseInterval(strings.Trim(args[1], `'`))
		if err != nil {
			return "", fmt.Errorf("error parsing interval %v", args[1])
		}
		if len(args) == 3 {
			err := SetupFillmode(configStruct, args[2])
			if err != nil {
				return "", err
			}
		}

		return fmt.Sprintf("TIME_SLICE(TO_TIMESTAMP(%s), %v, 'SECOND', 'START')", args[0], interval.Seconds()), nil
	case "__timeGroupAlias":
		tg, err := evaluateMacro("__timeGroup", args, configStruct)
		if err == nil {
			return tg + " AS time", nil
		}
		return "", err
	case "__unixEpochFilter":
		if len(args) == 0 {
			return "", fmt.Errorf("missing time column argument for macro %v", name)
		}
		return fmt.Sprintf("%s >= %d AND %s <= %d", args[0], timeRange.From.UTC().Unix(), args[0], timeRange.To.UTC().Unix()), nil
	case "__unixEpochNanoFilter":
		if len(args) == 0 {
			return "", fmt.Errorf("missing time column argument for macro %v", name)
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
		interval, err := ParseInterval(strings.Trim(args[1], `'`))
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
