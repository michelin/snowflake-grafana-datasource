package main

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestContainsIgnoreCase(t *testing.T) {

	tcs := []struct {
		array   []string
		str     string
		success bool
	}{
		{array: []string{"value1", "value2"}, str: "VALUE1", success: true},
		{array: []string{"value1", "value2"}, str: "VALUE2", success: true},
		{array: []string{"value 1", "value2"}, str: "value 1", success: true},
		{array: []string{"value1", "value2"}, str: "", success: false},
		{array: []string{}, str: "", success: false},
	}
	for i, tc := range tcs {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			if tc.success {
				require.True(t, equalsIgnoreCase(tc.array, tc.str))
			} else {
				require.False(t, equalsIgnoreCase(tc.array, tc.str))
			}
		})
	}
}

func TestMin(t *testing.T) {

	tcs := []struct {
		val1   int64
		val2   int64
		result int64
	}{
		{val1: 1, val2: 2, result: 1},
		{val1: 2, val2: 2.0, result: 2},
		{val1: 2, val2: 1, result: 1},
	}
	for i, tc := range tcs {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			require.Equal(t, tc.result, Min(tc.val1, tc.val2))
		})
	}
}

func TestMax(t *testing.T) {

	tcs := []struct {
		val1   int64
		val2   int64
		result int64
	}{
		{val1: 1, val2: 2, result: 2},
		{val1: 2, val2: 2.0, result: 2},
		{val1: 2, val2: 1, result: 2},
	}
	for i, tc := range tcs {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			require.Equal(t, tc.result, Max(tc.val1, tc.val2))
		})
	}
}
