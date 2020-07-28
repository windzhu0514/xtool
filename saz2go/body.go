package saz2go

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"strings"

	"github.com/ChimeraCoder/gojson"
	"github.com/windzhu0514/gocrypto"
	"github.com/windzhu0514/xtool/config"
	"github.com/windzhu0514/xtool/saz2go/structassgin"
)

type Parser interface {
	Parse(body []byte) ([]byte, interface{}, string, error)
}

type bodyParser struct {
	cfg         config.Decrypt
	contentType string
	isRequest   bool
	isParse     bool
	methodMame  string

	bodyData []byte
}

func (p *bodyParser) Parse(body io.ReadCloser) (
	rawBody []byte, dataDefine interface{}, dataAssign string, err error) {

	data, err := ioutil.ReadAll(body)
	if err != nil {
		body.Close()
		return nil, nil, "", err
	}

	body.Close()

	if len(data) == 0 {
		return nil, nil, "", nil
	}

	var ok bool
	if data, ok = fromBase64(data); ok {
		if p.cfg.AlgoName != "" {
			data, err = decrypt(p.cfg, data)
			if err != nil {
				return nil, nil, "", err
			}
		}
	}

	p.bodyData = data

	parser := p.ChooseParser()
	return parser.Parse(data)
}

func (p *bodyParser) ChooseParser() Parser {
	if strings.Contains(p.contentType, "application/json") || json.Valid(p.bodyData) {
		structName := p.methodMame
		if p.isRequest {
			structName += "Data"
		} else {
			structName += "Resp"
		}
		return jsonBodyParser{
			structName: structName,
		}
	}

	if strings.Contains(p.contentType, "application/x-www-form-urlencoded") {
		return formURLEncodeParser{
			cfg:     p.cfg,
			isParse: p.isParse,
		}
	}

	return rawParser{}
}

func fromBase64(in []byte) ([]byte, bool) {
	out := make([]byte, base64.StdEncoding.DecodedLen(len(in)))
	n, err := base64.StdEncoding.Decode(out, in)
	if err != nil {
		return in, false
	}
	return out[:n], true
}

type rawParser struct {
}

func (p rawParser) Parse(body []byte) ([]byte, interface{}, string, error) {
	return body, nil, "", nil
}

type formURLEncodeParser struct {
	cfg     config.Decrypt
	isParse bool
}

func (p formURLEncodeParser) Parse(body []byte) ([]byte, interface{}, string, error) {
	params, err := url.ParseQuery(string(body))
	if err != nil {
		return body, nil, "", err
	}

	formData := make([]string, 0, len(params))
	for k := range params {
		data := k
		v := params.Get(k)
		if vv, ok := fromBase64([]byte(v)); ok {
			if p.cfg.AlgoName != "" {
				// 解密form data
				vv, err = decrypt(p.cfg, vv)
				if err != nil {
					fmt.Println(err)
					return nil, nil, "", err
				}
			}

			v = string(vv)
		} else {
			v, _ = url.QueryUnescape(v)
		}

		data += "=" + v

		formData = append(formData, data)
	}

	if p.isParse {
		return []byte(strings.Join(formData, "&")), nil, "", nil
	}

	return body, "", "", nil
}

type jsonBodyParser struct {
	isParse    bool
	structName string
}

func (p jsonBodyParser) Parse(body []byte) ([]byte, interface{}, string, error) {
	if p.isParse {
		return body, nil, "", nil
	}

	structDefine, err := gojson.Generate(bytes.NewReader(body), gojson.ParseJson, p.structName, "main", []string{"json"}, false, true)
	if err != nil {
		return body, nil, "", err
	}

	structDefine = structDefine[bytes.Index(structDefine, []byte("type")):]
	structAssign, err := structassgin.Generate(body, p.structName)
	if err != nil {
		return body, nil, "", err
	}

	return body, string(structDefine), string(structAssign), nil
}

func decrypt(cfg config.Decrypt, data []byte) ([]byte, error) {
	ci := gocrypto.NewCipher(cfg.AlgoName)
	return ci.DecryptWithIV([]byte(cfg.Key), []byte(cfg.IV), data)
}
