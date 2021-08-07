package main

import "log"

type MaoLogLevel uint8

const (
	DEBUG MaoLogLevel = 0
	INFO MaoLogLevel = 1
	WARN MaoLogLevel = 2
	ERROR MaoLogLevel = 3
)

var (
	MaoLogLevelString = [4]string{"DEBUG", "INFO ", "WARN ", "ERROR"}
)

func MaoLog(level MaoLogLevel, logStr string) {
	switch level {
	case DEBUG:

	case INFO:
		fallthrough
	case WARN:
		fallthrough
	case ERROR:
		log.Printf("%s: %s", MaoLogLevelString[level], logStr)
	}
}