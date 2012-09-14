package datadog

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestPointArraySerialze(t *testing.T) {
	pa := PointArray{
		IsCount:      true,
		CountContent: []uint64{9999, 10001},
		GaugeContent: []float64{123.45, 12999.87},
		Timestamps:   []int64{1234, 1235},
	}
	{
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.Encode(&pa)
		result := strings.TrimSpace(string(buf.Bytes()))
		expected := "[[1234,9999],[1235,10001]]"
		if result != expected {
			t.Errorf("'%s': expected '%s'", result, expected)
		}
	}

	{
		pa.IsCount = false
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.Encode(&pa)
		result := strings.TrimSpace(string(buf.Bytes()))
		expected := "[[1234,123.450000],[1235,12999.870000]]"
		if result != expected {
			t.Errorf("'%s': expected '%s'", result, expected)
		}
	}
}
