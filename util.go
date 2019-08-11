package main

import "fmt"

func toInt64(val interface{}) (int64, error) {
	n, ok := val.(int64)
	if !ok {
		return 0, fmt.Errorf("Failed to cast %v (%T) to int64", val, val)
	}

	return n, nil
}

func boolPtr(val bool) *bool {
	return &val
}
