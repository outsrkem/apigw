package common

import "time"

func CreateTimestamp() int64 {
	t := time.Now().UnixNano() / 1e6
	return t
}
