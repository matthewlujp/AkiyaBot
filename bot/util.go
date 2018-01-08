package main

import (
	"fmt"
	"path"
	"strconv"
	"time"
)

func pathFromDateTime(now time.Time) []string {
	year, month, day := now.Date()
	return []string{
		"YasaiLog",
		strconv.Itoa(year),
		month.String(),
		fmt.Sprintf("%dth", day),
		fmt.Sprintf("%02d:%02d:%02d", now.Hour(), now.Minute(), now.Second()),
	}
}

func saveDirFromTime(now time.Time) string {
	year, month, day := now.Date()
	return path.Join(
		cnf.PhotoService.SaveDir, strconv.Itoa(year), month.String(), fmt.Sprintf("%dth", day),
		fmt.Sprintf("%02d%02d%02d", now.Hour(), now.Minute(), now.Second()),
	)
}
