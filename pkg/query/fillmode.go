package query

import "github.com/grafana/grafana-plugin-sdk-go/data"

// Constant used to describe the time series fill mode if no value has been seen
const (
	NullFill     = "null"
	PreviousFill = "previous"
	ValueFill    = "value"
)

func MapFillMode(fillModeString string) data.FillMode {
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
