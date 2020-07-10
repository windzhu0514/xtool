package saz2go

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/ChimeraCoder/gojson"

	"github.com/windzhu0514/xtool/config"
	"github.com/windzhu0514/xtool/crypto"
)

type Parser interface {
	Parse(body []byte) (string, string, error)
}

func parseBody(body io.Reader, isRequest bool) (string, string, error) {
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return "", "", err
	}

	data = fromBase64(data)

	parser := chooseParser(data, isRequest)
	return parser.Parse(data)

	// 默认明文解析 配置的加解密则进行对应的加解密
	// data, err = fromBase64(data)
	// if err != nil {
	// 	return err
	// }
}

func chooseParser(body []byte, isRequest bool) Parser {
	if json.Valid(body) {
		return jsonBodyParser{}
	}

	decryptCfg := config.Cfg.SAZ2go.Request.Decrypt
	if !isRequest {
		decryptCfg = config.Cfg.SAZ2go.Response.Decrypt
	}

	return cryptionParser{}
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

type formURLEncodeParser struct {
}

type jsonBodyParser struct {
}

func (p jsonBodyParser) Parse(body []byte) (string, string, error) {
	structDefine, err := gojson.Generate(bytes.NewReader(body), gojson.ParseJson, "name", "main", []string{"json"}, false, true)
	if err != nil {
		fmt.Println(err)
	}

	structDefine = structDefine[bytes.Index(structDefine, []byte("type")):]

	structAssign := ""

	return string(structDefine), structAssign, nil
}

type cryptionParser struct {
	decryptCfg config.Decrypt
}

func (p cryptionParser) Parse(body []byte) (string, string, error) {
	plainTxt, err := crypto.Decrypt(p.decryptCfg.AlgoName, body)
	if err != nil {
		return "", "", err
	}
}
