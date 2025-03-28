package utils

import "time"

func TimestampFormat(data int) string {
	timestamp := int64(data) // 示例时间戳（UTC时间2023-03-28 00:00:00）
	t := time.Unix(timestamp, 0)
	loc, _ := time.LoadLocation("Asia/Shanghai")
	return t.In(loc).Format("2006-01-02 15:04:05")
}
