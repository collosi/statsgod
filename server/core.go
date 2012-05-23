package server

type Datum struct {
	Value float64 `json:"y"`
	T     int64   `json:"x"`
}

type StatRecord struct {
	Name      string
	IsCounter bool
	Values    []Datum

	Capacity int
	w        int
}

type Core struct {
	Stats map[string]*StatRecord
}

type StatUpdate struct {
	Key       string
	Value     float64
	Timestamp int64
}

var (
	DefaultCapacity = 60
)

func NewCore() *Core {
	return &Core{make(map[string]*StatRecord)}
}

func (c *Core) Loop(updateChan chan StatUpdate, funcChan chan func(c *Core)) {
	for {
		select {
		case up, ok := <-updateChan:
			if !ok {
				return
			}
			c.Update(up)
		case f, ok := <-funcChan:
			if !ok {
				return
			}
			f(c)
		}

	}
}

func (c *Core) Update(up StatUpdate) {

	rec, ok := c.Stats[up.Key]
	if !ok {
		rec = &StatRecord{Capacity: DefaultCapacity}
		c.Stats[up.Key] = rec
	}
	rec.AppendValue(up.Value, up.Timestamp)
}

func (sr *StatRecord) CopyValues(arr *[]Datum) {
	*arr = make([]Datum, len(sr.Values))
	if sr.w == 0 {
		copy(*arr, sr.Values)
	} else {
		copy(*arr, sr.Values[sr.w:])
		copy((*arr)[len(sr.Values)-sr.w:], sr.Values[:sr.w])
	}
}

func (sr *StatRecord) AppendValue(f float64, t int64) {
	if sr.Capacity == 0 {
		sr.Capacity = DefaultCapacity
	}

	if len(sr.Values) < sr.Capacity {
		if len(sr.Values) == 0 {
			sr.Values = append(sr.Values, Datum{f, t})
			return
		}
		at := len(sr.Values) - 1
		if sr.Values[at].T == t {
			if sr.IsCounter {
				sr.Values[at].Value += f
			} else {
				sr.Values[at].Value = f
			}
		} else {
			sr.Values = append(sr.Values, Datum{f, t})
		}
	} else {
		previous := sr.w - 1
		if sr.w == 0 {
			previous = len(sr.Values) - 1
		}
		if sr.Values[previous].T == t {
			if sr.IsCounter {
				sr.Values[previous].Value += f
			} else {
				sr.Values[previous].Value = f
			}
		} else {
			sr.Values[sr.w].Value = f
			sr.Values[sr.w].T = t
			sr.w++
			sr.w = sr.w % len(sr.Values)
		}
	}
}
