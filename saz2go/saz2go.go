// TODO:支持json转结构体 生成赋值代码
package saz2go

import (
	"archive/zip"
	"bufio"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/PuerkitoBio/goquery"

	"github.com/windzhu0514/xtool/config"
	"github.com/windzhu0514/xtool/crypto"
)

var errInvalidFormat = errors.New("invalid fiddler saz file format")

type saz2go struct {
	structName      string
	structFirstChar string
	outputFileName  string
	tmplFileName    string
}

var ss = saz2go{}

func (s *saz2go) Convert(fileName string) error {
	pack, err := s.parse(fileName, false)
	if err != nil {
		return err
	}

	var t *template.Template
	if s.tmplFileName != "" {
		t, err = template.ParseFiles(s.tmplFileName)
	} else {
		t = template.New("req")
		t, err = t.Parse(tmplPackage)
	}

	if err != nil {
		return err
	}

	outputFileName := ss.outputFileName
	if outputFileName == "" {
		outputFileName = strings.Replace(fileName, ".saz", ".go", -1)
	} else {
		if !strings.HasSuffix(outputFileName, ".go") {
			outputFileName += ".go"
		}
	}

	f, err := os.OpenFile(outputFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	if err := t.Execute(f, pack); err != nil {
		return err
	}

	f.Close()

	return nil
}

func (s *saz2go) Parse(fileName string) error {
	pack, err := s.parse(fileName, true)
	if err != nil {
		return err
	}

	return s.save2file(pack, fileName)
}

func (s *saz2go) parse(fileName string, isParse bool) (*onePackage, error) {
	r, err := zip.OpenReader(fileName)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	files := make(map[string]*zip.File)
	for _, f := range r.File {
		files[f.Name] = f // raw/02_s.txt
	}

	indexFile, exist := files["_index.htm"]
	if !exist {
		return nil, errInvalidFormat
	}

	reader, err := indexFile.Open()
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, err
	}

	var pack onePackage
	pack.PackageName = "saz2go"
	pack.StructName = s.structName
	pack.StructNameFirstChar = s.structFirstChar

	var parseError error
	doc.Find("body table tbody tr").EachWithBreak(func(i int, ss *goquery.Selection) bool {
		reqFileName, ok0 := ss.Find("td").Eq(0).Find("a").Eq(0).Attr("href")
		respFileName, ok1 := ss.Find("td").Eq(0).Find("a").Eq(1).Attr("href")
		if ok0 && ok1 {
			reqFileName = strings.ReplaceAll(reqFileName, "\\", "/")
			respFileName = strings.ReplaceAll(respFileName, "\\", "/")

			var m oneMethod
			reqFile, exist := files[reqFileName]
			if exist {
				reqReader, err := reqFile.Open()
				if err != nil {
					parseError = err
					return false
				}

				err = s.parseRequest(&m, reqReader, i, isParse)
				if err != nil {
					parseError = err
					return false
				}

				m.StructNameFirstChar = pack.StructNameFirstChar
				m.StructName = pack.StructName

				reqReader.Close()
			}

			respFile, exist := files[respFileName]
			if exist {
				respReader, err := respFile.Open()
				if err != nil {
					parseError = err
					return false
				}

				err = s.parseResponse(&m, respReader, i, isParse)
				if err != nil {
					parseError = err
					return false
				}

				respReader.Close()
			}

			pack.Methods = append(pack.Methods, m)
		}

		return true
	})

	if parseError != nil {
		return nil, parseError
	}

	return &pack, nil
}

func (s *saz2go) save2file(pack *onePackage, fileName string) error {
	if pack == nil {
		return errors.New("package not have method")
	}

	var buf bytes.Buffer
	for _, v := range pack.Methods {
		buf.WriteString(v.URL)
		buf.WriteString("\n\n")
		buf.WriteString(v.ReqBody)
		buf.WriteString("\n\n")
		buf.WriteString(v.RespBody)
		buf.WriteString("\n\n\n")
	}

	fileName = strings.TrimSuffix(fileName, ".saz")
	fileName += ".txt"

	saveFile, err := os.OpenFile(fileName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	saveFile.Write(buf.Bytes())
	saveFile.Close()

	return nil
}

func (s *saz2go) parseRequest(m *oneMethod, r io.Reader, methodIndex int, isParse bool) error {
	request, err := http.ReadRequest(bufio.NewReader(r))
	if err != nil {
		return err
	}

	m.URL = request.URL.String()
	data, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return err
	}

	data = fromBase64(data)
	decryptCfg := config.Cfg.SAZ2go.Request.Decrypt
	if decryptCfg.AlgoName != "" {
		data, err = crypto.Decrypt(decryptCfg, data)
		if err != nil {
			return nil
		}
	}

	rawBody, dataDefine, dataAssign, err := parseBody(data, true)
	if err != nil {
		return err
	}

	m.ReqBody = string(rawBody)
	m.ReqDataDefine = dataDefine
	m.ReqDataAssign = dataAssign

	if isParse {
		return nil
	}

	m.Params = request.URL.Query()
	m.MethodMame = s.methodName(methodIndex, request.URL.Path)
	m.HttpMethod = strings.Title(strings.ToLower(request.Method))

	m.Heads = request.Header
	m.Heads.Add("Host", request.URL.Host)
	delete(m.Heads, "Cookie")
	delete(m.Heads, "Content-Length")
	delete(m.Heads, "Dnt")
	delete(m.Heads, "Upgrade-Insecure-Requests")

	for _, c := range config.Cfg.SAZ2go.Cookie.Remove {
		delete(m.Heads, c)
	}

	return nil
}

func (s *saz2go) parseResponse(m *oneMethod, r io.Reader, methodIndex int, isParse bool) error {
	response, err := http.ReadResponse(bufio.NewReader(r), nil)
	if err != nil {
		return err
	}
	for _, v := range response.Header.Values("Content-Type") {
		if strings.Contains(v, "text/html") {
			return nil
		}
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	data = fromBase64(data)
	decryptCfg := config.Cfg.SAZ2go.Response.Decrypt
	if decryptCfg.AlgoName != "" {
		data, err = crypto.Decrypt(decryptCfg, data)
		if err != nil {
			return nil
		}
	}

	m.RespBody = string(data)

	if isParse {
		return nil
	}

	return nil
}

// 正斜杠后面的单词作为方法名，首字母小写
func (s *saz2go) methodName(methodIndex int, path string) string {
	path = strings.TrimSuffix(path, "/")
	index := strings.LastIndex(path, "/")
	if index > 0 {
		name := path[index:]
		if len(name) > 0 {
			name = strings.ToLower(name[:1]) + name[1:]
		}
		return name
	}

	return "defaultMethod" + strconv.Itoa(methodIndex)
}

type onePackage struct {
	PackageName         string
	StructName          string
	StructNameFirstChar string
	Methods             []oneMethod
}

type oneMethod struct {
	StructNameFirstChar string
	StructName          string
	MethodMame          string
	HttpMethod          string
	URL                 string
	Heads               http.Header
	Params              url.Values

	ReqBody        string // 经过解密的原始body
	RespBody       string // 经过解密的原始body
	ReqDataDefine  interface{}
	ReqDataAssign  string
	RespDataDefine interface{}
	RespDataAssign string
}

var tmplPackage = `
package {{.PackageName}}

type {{.StructName}} struct {
	
}

{{range .Methods}}
func ({{.StructNameFirstChar}} *{{.StructName}}) {{.MethodMame}}() (resp string, err error) {
	for i := 0; i < conf.GSystemConfig.ReTryTimes; i++ {
        {{if .ReqMethod -}}
		req := httpclient.{{.ReqMethod}}("{{.URL}}")
		{{- end}}
		{{- range $key, $value :=  .Heads}}
		req.Header("{{$key}}", "{{index $value 0}}")
		{{- end}}
	    {{if .Params}}
		{{- range $key, $value :=  .Params}}
		req.Param("{{$key}}", "{{index $value 0}}")
		{{- end}}
		{{- end}}
		
		req.SetCookieJar({{.StructNameFirstChar}}.ci.CICookieJar)

        oldProxyIP := ""
		if {{.StructNameFirstChar}}.ci.UseProxy {
			proxyInfo, status := proxyMgr.GProxyMgr.Get()
			oldProxyIP = proxyInfo.ProxyIP
			if status {
				req.SetAuthProxy(proxyInfo.ProxyUser, proxyInfo.ProxyPass, proxyInfo.ProxyIP, proxyInfo.ProxyPort)
				airlog.GSLog.Info({{.StructNameFirstChar}}.logPrefix+" siteid=%d, ProxyID=%s, proxyip=%s:[采用新版代理]",
					common.AIR9C, proxyInfo.ProxyID, proxyInfo.ProxyIP)
			} else {
				airlog.GSLog.Info({{.StructNameFirstChar}}.logPrefix + "新版代理获取失败")
				utils.WaitRandMs(300, 500)
				continue
			}
		}
		resp, err = req.String()
		code, _ := req.GetStatusCode()
		if code >= http.StatusBadRequest || utils.IsProxyTimeout(err) {
			airlog.GSLog.Debug(l.logPrefix+" 返回结果异常 statuscode:%d", code)
			utils.WaitRandMs(1500, 2000)
			if {{.StructNameFirstChar}}.ci.UseProxy {
				proxyMgr.GProxyMgr.RedialProxyIP(oldProxyIP)
			}
			continue
		}

		if err != nil {
			airlog.GSLog.Error({{.StructNameFirstChar}}.logPrefix+"{{.MethodMame}} 请求失败：%+v", err.Error())
			utils.WaitRandMs(1500, 2000)
			continue
		}

		break
	}

	return resp, err
}	
{{end}}
`
