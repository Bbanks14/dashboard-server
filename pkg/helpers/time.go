package helpers

import (
	"fmt"
	"time"
)

// customTimeLayout creates a custom time layout of "DD-MM-YYYY HH-MM-SS"
const customTimeLayout = "02-01-2006 15:04:05"

// customTime wraps time.Time to use a custom JSON format
type customTime struct {
	Time    time.Time
	IsValid bool
}

// MarshalJSON converts time to custom format
func (ct customTime) MarshalJSON() ([]byte, error) {
	if !ct.IsValid {
		return []byte("null"), nil
	}
	return []byte(`"` + ct.Time.Format(customTimeLayout) + `"`), nil
}

// UnmarshalJSON parses custom format into time.Time
func (ct *customTime) UnmarshalJSON(data []byte) error {
	s := string(data)

	if s == "null" {
		ct.IsValid = false
		return nil
	}

	t, err := time.Parse(`"`+customTimeLayout+`"`, s)
	if err != nil {
		return fmt.Errorf("parsing customTime: %w", err)
	}

	ct.Time = t
	ct.IsValid = true
	return nil
}

// Use customTime in your struct
// type Event struct {
// Timestamp customTime `json:timestamp"`
// }
