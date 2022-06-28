package util

import (
	"log"
	"testing"
)

func TestInitMaoLog(t *testing.T) {
	log.Println(ERROR > DEBUG)
	log.Println(WARN > DEBUG)
	log.Println(INFO > DEBUG)
	log.Println(DEBUG > DEBUG)
}