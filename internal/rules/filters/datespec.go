package filters

import (
	"fmt"
	"time"
)

// DateSpec represents a point in time, either relative or absolute.
// Exactly one field should be set.
type DateSpec struct {
	SecondsAgo *float64 `yaml:"seconds_ago"`
	MinutesAgo *float64 `yaml:"minutes_ago"`
	HoursAgo   *float64 `yaml:"hours_ago"`
	DaysAgo    *float64 `yaml:"days_ago"`
	WeeksAgo   *float64 `yaml:"weeks_ago"`
	MonthsAgo  *float64 `yaml:"months_ago"`
	YearsAgo   *float64 `yaml:"years_ago"`
	Unix       *int64   `yaml:"unix"`
	Date       *string  `yaml:"date"`
}

// ToTime converts the DateSpec to an absolute time.Time value.
func (d *DateSpec) ToTime() (time.Time, error) {
	now := time.Now()

	switch {
	case d.SecondsAgo != nil:
		return now.Add(-time.Duration(*d.SecondsAgo * float64(time.Second))), nil
	case d.MinutesAgo != nil:
		return now.Add(-time.Duration(*d.MinutesAgo * float64(time.Minute))), nil
	case d.HoursAgo != nil:
		return now.Add(-time.Duration(*d.HoursAgo * float64(time.Hour))), nil
	case d.DaysAgo != nil:
		return now.Add(-time.Duration(*d.DaysAgo * 24 * float64(time.Hour))), nil
	case d.WeeksAgo != nil:
		return now.Add(-time.Duration(*d.WeeksAgo * 7 * 24 * float64(time.Hour))), nil
	case d.MonthsAgo != nil:
		// Approximate: 30.44 days per month
		return now.Add(-time.Duration(*d.MonthsAgo * 30.44 * 24 * float64(time.Hour))), nil
	case d.YearsAgo != nil:
		// Approximate: 365.25 days per year
		return now.Add(-time.Duration(*d.YearsAgo * 365.25 * 24 * float64(time.Hour))), nil
	case d.Unix != nil:
		return time.Unix(*d.Unix, 0), nil
	case d.Date != nil:
		// Try parsing as YYYY-MM-DDTHH:MM:SS first, then YYYY-MM-DD
		t, err := time.Parse("2006-01-02T15:04:05", *d.Date)
		if err != nil {
			t, err = time.Parse("2006-01-02", *d.Date)
			if err != nil {
				return time.Time{}, fmt.Errorf("invalid date format %q: expected YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS", *d.Date)
			}
		}
		return t, nil
	default:
		return time.Time{}, fmt.Errorf("no time specification provided")
	}
}

// Validate checks that exactly one time specification is set.
func (d *DateSpec) Validate() error {
	count := 0
	if d.SecondsAgo != nil {
		count++
	}
	if d.MinutesAgo != nil {
		count++
	}
	if d.HoursAgo != nil {
		count++
	}
	if d.DaysAgo != nil {
		count++
	}
	if d.WeeksAgo != nil {
		count++
	}
	if d.MonthsAgo != nil {
		count++
	}
	if d.YearsAgo != nil {
		count++
	}
	if d.Unix != nil {
		count++
	}
	if d.Date != nil {
		count++
	}

	if count == 0 {
		return fmt.Errorf("no time specification provided")
	}
	if count > 1 {
		return fmt.Errorf("only one time specification allowed")
	}

	// Validate unix timestamp is positive
	if d.Unix != nil && *d.Unix < 0 {
		return fmt.Errorf("unix timestamp must be positive")
	}

	return nil
}
