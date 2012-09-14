package datadog

import (
	"bytes"
	"fmt"
	"io"
)

type PointArray struct {
	IsCounter    bool
	CountContent []uint64
	GaugeContent []float64
	Timestamps   []int64
}

type Series struct {
	Metric string     `json:"metric"`
	Points PointArray `json:"points"`
	Type   string     `json:"type"`
	Host   string     `json:"host"`
	Device string     `json:"device"`
	Tags   []string   `json:"tags"`
}

func NewCountPointArray(size int) PointArray {
	return PointArray{
		IsCounter:    true,
		CountContent: make([]uint64, size),
		Timestamps:   make([]int64, size),
	}
}

func NewGaugePointArray(size int) PointArray {
	return PointArray{
		IsCounter:    false,
		GaugeContent: make([]float64, size),
		Timestamps:   make([]int64, size),
	}
}
func (pa *PointArray) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	io.WriteString(&buf, "[")
	l := len(pa.Timestamps)
	var comma string
	if pa.IsCounter {
		for i := 0; i < l; i++ {
			if i > 0 {
				comma = ","
			}
			fmt.Fprintf(&buf, "%s[%d,%d]", comma, pa.Timestamps[i], pa.CountContent[i])
		}
	} else {
		for i := 0; i < l; i++ {
			if i > 0 {
				comma = ","
			}
			fmt.Fprintf(&buf, "%s[%d,%f]", comma, pa.Timestamps[i], pa.GaugeContent[i])

		}
	}

	io.WriteString(&buf, "]")
	return buf.Bytes(), nil
}
