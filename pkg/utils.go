package main

import (
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

func equalsIgnoreCase(s []string, str string) bool {
	for _, v := range s {
		if strings.EqualFold(v, str) {
			return true
		}
	}
	return false
}

// Max return the maximum value between two int64
func Max(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}

// Min return the minimum value between two int64
func Min(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

func insertFrameField(frame *data.Frame, value interface{}, index int) {
	switch v := value.(type) {
	case string:
		frame.Fields[index].Append(&v)
	case float64:
		frame.Fields[index].Append(&v)
	case int64:
		frame.Fields[index].Append(&v)
	case bool:
		frame.Fields[index].Append(&v)
	case time.Time:
		frame.Fields[index].Append(&v)
	default:
		frame.Fields[index].Append(nil)
	}
}

func previousRow(rows [][]interface{}, index int) []interface{} {
	if len(rows) > 0 {
		return rows[Max(int64(index-1), 0)]
	}
	return nil
}
