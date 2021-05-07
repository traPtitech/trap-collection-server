package values

import "time"

type nullableTime struct {
	Time time.Time
	IsNull bool
}

var (
	nullTime = nullableTime{
		IsNull: true,
	}
)
