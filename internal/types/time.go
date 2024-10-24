package types

import (
	"errors"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/jackc/pgx/v5/pgtype"
	"gopkg.in/guregu/null.v4"

	"github.com/thanishsid/dingilink-server/internal/types/apperror"
)

const TimeOfDayFormat = "15:04"

func ParseTimeOfDay(s string) (time.Time, error) {
	return time.Parse(TimeOfDayFormat, s)
}

func TimeToMicrosecondsFromMidnight(t time.Time) int64 {
	hours := t.Hour()
	minutes := t.Minute()

	hoursMicroseconds := hours * 3600000000
	minutesMicroseconds := minutes * 60000000

	return int64(hoursMicroseconds + minutesMicroseconds)
}

func MicrosecondsToTimeOfDayString(ms int64) (string, error) {
	t, err := time.Parse(TimeOfDayFormat, "00:00")
	if err != nil {
		return "", err
	}

	t = t.Add(time.Duration(ms) * time.Microsecond)

	return t.Format(TimeOfDayFormat), nil
}

func PgTimeToTimeOfDay(t pgtype.Time) (*string, error) {
	if t.Valid {
		tod, err := MicrosecondsToTimeOfDayString(t.Microseconds)
		return &tod, err
	}

	return nil, nil
}

func MustPgTimeToTimeOfDay(t pgtype.Time) *string {
	if t.Valid {
		tod, err := MicrosecondsToTimeOfDayString(t.Microseconds)
		if err != nil {
			panic("failed to convert microseconds to time of day: " + err.Error())
		}

		return &tod
	}

	return nil
}

var ValidateTimeOfDay = validation.By(func(value interface{}) error {
	var val string

	switch v := value.(type) {
	case string:
		val = v
	case *string:
		val = null.StringFromPtr(v).ValueOrZero()
	default:
		return errors.New(apperror.INPUT_INVALID)
	}

	_, err := ParseTimeOfDay(val)
	if err != nil {
		return errors.New(apperror.INPUT_INVALID) 
	}

	return nil
})
