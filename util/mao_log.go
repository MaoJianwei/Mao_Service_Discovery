package util

import (
	"fmt"
	"log"
	"os"
)

type MaoLogLevel uint8

const (
	DEBUG MaoLogLevel = 0
	INFO  MaoLogLevel = 1
	WARN  MaoLogLevel = 2
	ERROR MaoLogLevel = 3
)

var (
	MaoLogLevelString = [4]string{"DEBUG", "INFO ", "WARN ", "ERROR"}
)

func InitMaoLog() {
	log.SetOutput(os.Stdout)
}

func MaoLog(level MaoLogLevel, format string, a ...interface{}) {
	switch level {
	case DEBUG:
		fallthrough
	case INFO:
		fallthrough
	case WARN:
		fallthrough
	case ERROR:
		log.Printf("%s: %s", MaoLogLevelString[level], fmt.Sprintf(format, a...))
	}
}

func MaoLogM(level MaoLogLevel, moduleName string, format string, a ...interface{}) {
	switch level {
	case DEBUG:
		fallthrough
	case INFO:
		fallthrough
	case WARN:
		fallthrough
	case ERROR:
		log.Printf("%s: %s: %s", MaoLogLevelString[level], moduleName, fmt.Sprintf(format, a ...))
	}
}