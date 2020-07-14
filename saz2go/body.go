package saz2go

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/url"

	"github.com/ChimeraCoder/gojson"
)

type Parser interface {
	Parse(body []byte, isParse bool) ([]byte, interface{}, string, error)
}

func parseBody(body []byte, isParse bool) ([]byte, interface{}, string, error) {
	parser := chooseParser(body)
	return parser.Parse(body, isParse)
}

func chooseParser(body []byte) Parser {
	if json.Valid(body) {
		return jsonBodyParser{}
	}

	if _, err := url.ParseQuery(string(body)); err == nil {
		return formURLEncodeParser{}
	}

	return rawParser{}
}

func fromBase64(in []byte) []byte {
	out := make([]byte, base64.StdEncoding.DecodedLen(len(in)))
	_, err := base64.StdEncoding.Decode(out, in)
	if err != nil {
		return in
	}

	return out
}

type rawParser struct {
}

func (p rawParser) Parse(body []byte, isParse bool) ([]byte, interface{}, string, error) {
	return body, nil, "", nil
}

type formURLEncodeParser struct {
}

func (p formURLEncodeParser) Parse(body []byte, isParse bool) ([]byte, interface{}, string, error) {
	raw, err := url.QueryUnescape(string(body))
	if err != nil {
		return body, nil, "", err
	}

	if isParse {
		return []byte(raw), nil, "", err
	}

	params, err := url.ParseQuery(string(body))
	if err != nil {
		return []byte(raw), nil, "", err
	}

	return []byte(raw), params, "", nil
}

type jsonBodyParser struct {
}

func (p jsonBodyParser) Parse(body []byte, isParse bool) ([]byte, interface{}, string, error) {
	if isParse {
		return body, nil, "", nil
	}

	structDefine, err := gojson.Generate(bytes.NewReader(body), gojson.ParseJson, "name", "main", []string{"json"}, false, true)
	if err != nil {
		return body, nil, "", err
	}

	structDefine = structDefine[bytes.Index(structDefine, []byte("type")):]
	structAssign := ""
	return body, string(structDefine), structAssign, nil
}
