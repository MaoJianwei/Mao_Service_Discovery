package AuxDataProcessor

import (
	"encoding/json"
	"log"
	"testing"
)

func TestEnvTempProcessor_Process(t *testing.T) {
	mmm := make(map[string]interface{})
	mmm["envTemp"] = 26.588
	mmm["v6In"] = 0x12345678ABCDEF96
	mmm["v6Out"] = "Bigmao Radar"

	b, err := json.Marshal(mmm)
	if err != nil {
		log.Println(err.Error())
	} else {
		sss := string(b)
		log.Println(sss)
		log.Println(b)
	}
}