package utils

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseInterval(t *testing.T) {
	daysInMonth, daysInYear := calculateDays()

	tcs := []struct {
		inp      string
		duration time.Duration
		err      *regexp.Regexp
	}{
		{inp: "1d", duration: 24 * time.Hour},
		{inp: "1w", duration: 24 * 7 * time.Hour},
		{inp: "2w", duration: 24 * 7 * 2 * time.Hour},
		{inp: "1M", duration: time.Duration(daysInMonth * 24 * int(time.Hour))},
		{inp: "1y", duration: time.Duration(daysInYear * 24 * int(time.Hour))},
		{inp: "5y", duration: time.Duration(calculateDays5y() * 24 * int(time.Hour))},
		{inp: "1x", err: regexp.MustCompile(`time: unknown unit "x" in duration "1x"`)},
		{inp: "invalid-duration", err: regexp.MustCompile(`^time: invalid duration "?invalid-duration"?$`)},
		{inp: "", err: regexp.MustCompile(`^time: invalid duration ""`)},
	}
	for i, tc := range tcs {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			res, err := ParseInterval(tc.inp)
			if tc.err == nil {
				require.NoError(t, err, "input %q", tc.inp)
				require.Equal(t, tc.duration, res, "input %q", tc.inp)
			} else {
				require.Error(t, err, "input %q", tc.inp)
				require.Regexp(t, tc.err, err.Error())
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	tcs := []struct {
		inp      string
		duration time.Duration
		err      *regexp.Regexp
	}{
		{inp: "1s", duration: time.Second},
		{inp: "1m", duration: time.Minute},
		{inp: "1h", duration: time.Hour},
		{inp: "1d", duration: 24 * time.Hour},
		{inp: "1w", duration: 7 * 24 * time.Hour},
		{inp: "2w", duration: 2 * 7 * 24 * time.Hour},
		{inp: "1M", duration: time.Duration(730.5 * float64(time.Hour))},
		{inp: "1y", duration: 365.25 * 24 * time.Hour},
		{inp: "5y", duration: 5 * 365.25 * 24 * time.Hour},
		{inp: "invalid-duration", err: regexp.MustCompile(`^time: invalid duration "?invalid-duration"?$`)},
		{inp: "", err: regexp.MustCompile(`^time: invalid duration ""`)},
	}
	for i, tc := range tcs {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			res, err := ParseDuration(tc.inp)
			if tc.err == nil {
				require.NoError(t, err, "input %q", tc.inp)
				require.Equal(t, tc.duration, res, "input %q", tc.inp)
			} else {
				require.Error(t, err, "input %q", tc.inp)
				require.Regexp(t, tc.err, err.Error())
			}
		})
	}
}

func calculateDays() (int, int) {
	now := time.Now().UTC()
	currentYear, currentMonth, _ := now.Date()

	firstDayOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, time.UTC)
	daysInMonth := firstDayOfMonth.AddDate(0, 1, -1).Day()

	daysInYear := int(now.AddDate(1, 0, 0).Sub(now).Hours() / 24)

	return daysInMonth, daysInYear
}

func calculateDays5y() int {
	now := time.Now().UTC()

	daysInYear := int(now.AddDate(5, 0, 0).Sub(now).Hours() / 24)

	return daysInYear
}
