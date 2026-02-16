package scalars

import (
	"io"
	"time"

	"github.com/99designs/gqlgen/graphql"
)

// Time is a custom scalar for Date/Time
type Time time.Time

// UnmarshalGQL implements graphql.Unmarshaler
func (t *Time) UnmarshalGQL(v interface{}) error {
	if str, ok := v.(string); ok {
		parsed, err := time.Parse(time.RFC3339, str)
		if err != nil {
			return err
		}
		*t = Time(parsed)
		return nil
	}
	return nil
}

// MarshalGQL implements graphql.Marshaler
func (t Time) MarshalGQL(w io.Writer) {
	graphql.MarshalTime(time.Time(t)).MarshalGQL(w)
}
