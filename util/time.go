package util

import "time"

// NowMilli returns now in milliseconds since epoch
func NowMilli() int64 {
	return time.Now().UnixNano() / 1000000
}
