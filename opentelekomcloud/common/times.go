package common

import (
	"log"
	"time"
)

func FormatTimeStampRFC3339(timestamp int64, isUTC bool, customFormat ...string) string {
	if timestamp == 0 {
		return ""
	}

	createTime := time.Unix(timestamp, 0)
	if isUTC {
		createTime = createTime.UTC()
	}
	if len(customFormat) > 0 {
		return createTime.Format(customFormat[0])
	}
	return createTime.Format(time.RFC3339)
}

// ConvertTimeStrToNanoTimestamp is a method that used to convert the time string into the corresponding timestamp (in
// nanosecond), e.g.
// The supported time formats are as follows:
//   - RFC3339 format:
//     2006-01-02T15:04:05Z (default time format, if you are missing customFormat input)
//     2006-01-02T15:04:05.000000Z
//     2006-01-02T15:04:05Z08:00
//   - Other time formats:
//     2006-01-02 15:04:05
//     2006-01-02 15:04:05+08:00
//     2006-01-02T15:04:05
//     ...
//
// Two common uses are shown below:
// - ConvertTimeStrToNanoTimestamp("2024-01-01T00:00:00Z")
// - ConvertTimeStrToNanoTimestamp("2024-01-01T00:00:00+08:00", "2006-01-02T15:04:05Z08:00")
func ConvertTimeStrToNanoTimestamp(timeStr string, customFormat ...string) int64 {
	// The default time format is RFC3339.
	timeFormat := time.RFC3339
	if len(customFormat) > 0 {
		timeFormat = customFormat[0]
	}
	t, err := time.Parse(timeFormat, timeStr)
	if err != nil {
		log.Printf("error parsing the input time (%s), the time string does not match time format (%s): %s",
			timeStr, timeFormat, err)
		return 0
	}

	timestamp := t.UnixNano() / int64(time.Millisecond)
	// If the time is less than 1970-01-01T00:00:00Z, the timestamp is negative, such as: "0001-01-01T00:00:00Z"
	if timestamp < 0 {
		return 0
	}
	return timestamp
}
