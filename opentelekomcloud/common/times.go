package common

import "time"

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
