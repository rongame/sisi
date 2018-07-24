package util

import (
	"strconv"
	"time"
)

/**
 * 获取当前时间戳字符串
 */
func GetTimestampString() string {
	return strconv.FormatInt(GetTimestamp(), 10)
}

/**
 * 获取当前时间戳int64
 */
func GetTimestamp() int64 {
	return time.Now().Unix()
}

/**
 * 获取当前时间戳int
 */
func GetTimestampInt() int {
	timestamp, _ := strconv.Atoi(GetTimestampString())
	return timestamp
}

/**
 * 获取今天0点时间戳
 */
func GetTodayZeroInt() int {
	timeStr := time.Now().Format("2006-01-02")
	t, _ := time.Parse("2006-01-02", timeStr)
	timeNumber := t.Unix()
	timestamp, _ := strconv.Atoi(strconv.FormatInt(timeNumber, 10))

	return timestamp - 28800
}
