package values

import "time"

type nullableTime struct {
	time time.Time
	isNull bool
}

var (
	nullTime = nullableTime{
		isNull: true,
	}
)

func (nt *nullableTime) IsNull() bool {
	return nt.isNull
}

func (nt *nullableTime) Time() time.Time {
	return nt.time
}
