package Email

import (
	"fmt"
	"log"
	"testing"
)

func TestLoginAuth_Next(t *testing.T) {
	var fromServer []byte = []byte("Beijing")
	s := fmt.Sprintf("Unknown message from Server: %s", fromServer)
	log.Println(s)
}