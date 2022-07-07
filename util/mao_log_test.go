package util

import (
	"log"
	"testing"
)

func TestInitMaoLog(t *testing.T) {
	InitMaoLog(DEBUG)
	MaoLogM(DEBUG, "Test", "out %d", 666)
	MaoLogM(HOT_DEBUG, "Test", "out %d", 666)
	MaoLogM(INFO, "Test", "out %d", 666)
	MaoLogM(WARN, "Test", "out %d", 666)
	MaoLogM(ERROR, "Test", "out %d", 666)
	MaoLogM(SILENT, "Test", "out %d", 666)

	log.Println("==============================================")

	InitMaoLog(HOT_DEBUG)
	MaoLogM(DEBUG, "Test", "out %d", 666)
	MaoLogM(HOT_DEBUG, "Test", "out %d", 666)
	MaoLogM(INFO, "Test", "out %d", 666)
	MaoLogM(WARN, "Test", "out %d", 666)
	MaoLogM(ERROR, "Test", "out %d", 666)
	MaoLogM(SILENT, "Test", "out %d", 666)

	log.Println("==============================================")

	InitMaoLog(INFO)
	MaoLogM(DEBUG, "Test", "out %d", 666)
	MaoLogM(HOT_DEBUG, "Test", "out %d", 666)
	MaoLogM(INFO, "Test", "out %d", 666)
	MaoLogM(WARN, "Test", "out %d", 666)
	MaoLogM(ERROR, "Test", "out %d", 666)
	MaoLogM(SILENT, "Test", "out %d", 666)

	log.Println("==============================================")

	InitMaoLog(WARN)
	MaoLogM(DEBUG, "Test", "out %d", 666)
	MaoLogM(HOT_DEBUG, "Test", "out %d", 666)
	MaoLogM(INFO, "Test", "out %d", 666)
	MaoLogM(WARN, "Test", "out %d", 666)
	MaoLogM(ERROR, "Test", "out %d", 666)
	MaoLogM(SILENT, "Test", "out %d", 666)

	log.Println("==============================================")

	InitMaoLog(ERROR)
	MaoLogM(DEBUG, "Test", "out %d", 666)
	MaoLogM(HOT_DEBUG, "Test", "out %d", 666)
	MaoLogM(INFO, "Test", "out %d", 666)
	MaoLogM(WARN, "Test", "out %d", 666)
	MaoLogM(ERROR, "Test", "out %d", 666)
	MaoLogM(SILENT, "Test", "out %d", 666)

	log.Println("==============================================")

	InitMaoLog(SILENT)
	MaoLogM(DEBUG, "Test", "out %d", 666)
	MaoLogM(HOT_DEBUG, "Test", "out %d", 666)
	MaoLogM(INFO, "Test", "out %d", 666)
	MaoLogM(WARN, "Test", "out %d", 666)
	MaoLogM(ERROR, "Test", "out %d", 666)
	MaoLogM(SILENT, "Test", "out %d", 666)

	log.Println(SILENT > DEBUG)
	log.Println(ERROR > DEBUG)
	log.Println(WARN > DEBUG)
	log.Println(INFO > DEBUG)
	log.Println(HOT_DEBUG > DEBUG)
	log.Println(DEBUG > DEBUG)
}