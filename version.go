package main

import "time"

var (
	version   = time.Now().Local().Format("20060102150405")
	revision  string //nolint:gochecknoglobals
	buildDate string //nolint:gochecknoglobals
)
