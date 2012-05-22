package main

import (
	"testing"
)

func TestSimpleAdds(t *testing.T) {
	core := NewCore()
	core.Stats["v"] = &StatRecord{
		IsCounter: false,
		Capacity:  10,
	}
	core.Stats["c"] = &StatRecord{
		IsCounter: true,
		Capacity:  10,
	}

	// value type, add two dupes, then two distinct
	core.Update(StatUpdate{"v", 10.0, 1000})
	core.Update(StatUpdate{"v", 11.0, 1000})
	core.Update(StatUpdate{"v", 12.0, 1001})
	core.Update(StatUpdate{"v", 13.0, 1002})

	vr := core.Stats["v"]
	if len(vr.Values) != 3 {
		t.Errorf("%d: expected 3 values", len(vr.Values))
	}
	if vr.Values[0].T != 1000 || vr.Values[0].Value != 11.0 {
		t.Errorf("incorrect first value")
	}

	// counter type, add two dupes, then two distinct
	core.Update(StatUpdate{"c", 10.0, 1000})
	core.Update(StatUpdate{"c", 11.0, 1000})
	core.Update(StatUpdate{"c", 12.0, 1001})
	core.Update(StatUpdate{"c", 13.0, 1002})

	cr := core.Stats["c"]
	if len(cr.Values) != 3 {
		t.Errorf("%d: expected 3 values", len(cr.Values))
	}
	if cr.Values[0].T != 1000 || cr.Values[0].Value != 21.0 {
		t.Errorf("incorrect first value")
	}
	if cr.Values[1].Value != 12.0 {
		t.Errorf("%f: expected 12.0", cr.Values[2].Value)
	}
	if cr.Values[2].Value != 13.0 {
		t.Errorf("%f: expected 13.0", cr.Values[2].Value)
	}
}

func TestAddMoreThanCapacity(t *testing.T) {
	core := NewCore()
	core.Stats["v"] = &StatRecord{
		IsCounter: false,
		Capacity:  10,
	}
	core.Stats["c"] = &StatRecord{
		IsCounter: true,
		Capacity:  10,
	}

	for i := 1; i <= 50; i++ {
		core.Update(StatUpdate{"c", float64(i), int64(i)})
		core.Update(StatUpdate{"c", float64(i * 2), int64(i)})
		core.Update(StatUpdate{"v", float64(i), int64(i)})
		core.Update(StatUpdate{"v", float64(i * 2), int64(i)})

		var counts []Datum
		var vals []Datum
		core.Stats["c"].CopyValues(&counts)
		core.Stats["v"].CopyValues(&vals)

		if len(counts) != intMin(i, 10) {
			t.Errorf("%d: expected %d counts", len(counts), intMin(i, 10))
		}
		if counts[len(counts)-1].T != int64(i) || counts[len(counts)-1].Value != float64(3*i) {
			t.Errorf("%v: unexpected count entry at %d (%f)", counts[len(counts)-1], i, float64(3*i))
		}
		if len(vals) != intMin(i, 10) {
			t.Errorf("%d: expected %d vals", len(vals), intMin(i, 10))
		}
		if vals[len(vals)-1].T != int64(i) || vals[len(vals)-1].Value != float64(i*2) {
			t.Errorf("%v: unexpected value entry at %d", vals[len(vals)-1], i)
		}
	}
}

func intMin(i1, i2 int) int {
	if i1 < i2 {
		return i1
	}
	return i2
}
