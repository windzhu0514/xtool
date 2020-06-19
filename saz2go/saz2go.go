// TODO:支持所有功能模块的生成
// TODO:支持json转结构体 生成赋值代码
// TODO：生成站点工程 生成单个文件（请求流程） 只生成方法
package saz2go

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/ChimeraCoder/gojson"
	"github.com/PuerkitoBio/goquery"
)

var errInvalidFormat = errors.New("invalid fiddler saz file format")

type saz2go struct {
	structName      string
	structFirstChar string
	outputFileName  string
	tmplFileName    string
}

var ss = saz2go{}

func (s *saz2go) Run(fileName string) error {
	r, err := zip.OpenReader(fileName)
	if err != nil {
		return err
	}
	defer r.Close()

	var pack onePackage
	pack.PackageName = "packagename"
	pack.StructName = s.structName
	pack.StructNameFirstChar = s.structFirstChar

	files := make(map[string]*zip.File)
	for _, f := range r.File {
		files[f.Name] = f // raw/02_s.txt
	}

	indexFile, exist := files["_index.htm"]
	if !exist {
		return errInvalidFormat
	}

	read, err := indexFile.Open()
	if err != nil {
		return err
	}

	doc, err := goquery.NewDocumentFromReader(read)
	if err != nil {
		return err
	}

	var parseError error
	doc.Find("body table tbody tr").EachWithBreak(func(i int, ss *goquery.Selection) bool {

		reqName, ok0 := ss.Find("td a").Eq(0).Attr("href")
		respName, ok1 := ss.Find("td a").Eq(1).Attr("href")

		if ok0 && ok1 {
			reqName = strings.Replace(reqName, "\\", "/", -1)
			respName = strings.Replace(respName, "\\", "/", -1)

			reqFile, exist := files[reqName]
			if exist {
				reqRead, err := reqFile.Open()
				if err != nil {
					parseError = err
					return false
				}

				method, err := s.parseRequest(i, bufio.NewReader(reqRead))
				if err != nil {
					parseError = err
					return false
				}

				method.StructNameFirstChar = pack.StructNameFirstChar
				method.StructName = pack.StructName
				pack.Methods = append(pack.Methods, method)

				reqRead.Close()
			}
		}

		return true
	})

	if parseError != nil {
		return parseError
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

func (s *saz2go) parseRequest(index int, rc *bufio.Reader) (oneMethod, error) {
	request, err := http.ReadRequest(rc)
	if err != nil {
		return oneMethod{}, err
	}

	var m oneMethod
	m.RetryTimes = 3

	m.URL = request.URL.String()
	m.Params=request.URL.Query()
	// 正斜杠后面的单词作为方法名 大小写统一为title风格
	m.MethodMame = "defaultMethod" + strconv.Itoa(index)
	m.ReqMethod = strings.Title(strings.ToLower(request.Method))

	m.Heads = request.Header
	delete(m.Heads, "Cookie")
	delete(m.Heads, "Content-Length")
	delete(m.Heads, "Dnt")
	delete(m.Heads, "Upgrade-Insecure-Requests")

	body, err := ioutil.ReadAll(request.Body)
	if err == nil {
		m.Body = string(body)
	}
	if json.Valid(body){
		m.IsJson=true
		s, err := gojson.Generate(bytes.NewReader(body), gojson.ParseJson, "name", "main", []string{"json"}, false, true)
		if err!=nil{
			fmt.Println(err)
		}
		s=s[bytes.Index(s,[]byte("type")):]
	}

	contentTypes := m.Heads.Values("Content-Type")
	for _, ct := range contentTypes {
		if strings.Contains(ct,"application/x-www-form-urlencoded") {
			m.Params, err = url.ParseQuery(m.Body)
			if err != nil {
				return m, err
			}
		}
	}


	return m, nil
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
	RetryTimes          int
	ReqMethod           string
	URL                 string
	Heads               http.Header
	Params              url.Values

	Body                string
	IsJson bool
	IsForm bool
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
					common.CSAIR, proxyInfo.ProxyID, proxyInfo.ProxyIP)
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
