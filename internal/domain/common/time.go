package common

import (
	"fmt"
	"time"
)

type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

func NewTimeRange(start, end time.Time) (TimeRange, error) {
	start = start.UTC()
	end = end.UTC()
	if start.IsZero() || end.IsZero() {
		return TimeRange{}, FieldError("time", "start and end are required")
	}
	if !end.After(start) {
		return TimeRange{}, FieldError("time", "end must be after start")
	}
	return TimeRange{Start: start, End: end}, nil
}

func MustTimeRange(start, end time.Time) TimeRange {
	r, err := NewTimeRange(start, end)
	if err != nil {
		panic(fmt.Sprintf("invalid time range: %v", err))
	}
	return r
}

func (r TimeRange) Overlaps(other TimeRange) bool {
	return r.Start.Before(other.End) && other.Start.Before(r.End)
}

func (r TimeRange) Contains(at time.Time) bool {
	at = at.UTC()
	return !at.Before(r.Start) && at.Before(r.End)
}

func (r TimeRange) Duration() time.Duration { return r.End.Sub(r.Start) }
