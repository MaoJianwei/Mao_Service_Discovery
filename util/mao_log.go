package util

import (
	"fmt"
	"log"
	"os"
)

type MaoLogLevel uint8

const (
	DEBUG MaoLogLevel = iota
	HOT_DEBUG
	INFO
	WARN
	ERROR
	SILENT
)

var (
	MaoLogLevelString = [6]string{"DEBUG", "HOT_DEBUG", "INFO ", "WARN ", "ERROR", "SILENT"}
	minShowingLevel   = INFO // default is INFO
)

func InitMaoLog(minLogLevel MaoLogLevel) {
	log.SetOutput(os.Stdout)
	minShowingLevel = minLogLevel
}

func MaoLog(level MaoLogLevel, format string, a ...interface{}) {
	if minShowingLevel > level || minShowingLevel == SILENT {
		return
	}
	log.Printf("%s: %s", MaoLogLevelString[level], fmt.Sprintf(format, a...))
}

func MaoLogM(level MaoLogLevel, moduleName string, format string, a ...interface{}) {
	if minShowingLevel > level || minShowingLevel == SILENT {
		return
	}
	log.Printf("%s: %s: %s", MaoLogLevelString[level], moduleName, fmt.Sprintf(format, a ...))
}
