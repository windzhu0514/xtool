package saz2go

import (
	"encoding/base64"
	"fmt"
	"testing"
)

func TestBase64(t *testing.T) {
	str := "111111"
	fmt.Println(base64.StdEncoding.DecodeString(str))
}

func TestJsonParser(t *testing.T) {
	var j jsonBodyParser
	fmt.Println(j.Parse([]byte(`{"doNotRetry": false,"engineType": "web","robotId": 10041}`)))
}
