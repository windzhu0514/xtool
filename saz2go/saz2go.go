// TODO:支持所有功能模块的生成
// TODO:支持json转结构体 生成赋值代码
package saz2go

import (
	"archive/zip"
	"bufio"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"text/template"

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
	m.MethodMame = "defaultMethod" + strconv.Itoa(index)
	m.ReqMethod = request.Method
	m.Heads = request.Header
	delete(m.Heads, "Cookie")
	delete(m.Heads, "Content-Length")

	body, err := ioutil.ReadAll(request.Body)
	if err == nil {
		m.Body = string(body)
	}

	contentType := m.Heads.Get("Content-Type")
	if contentType == "application/x-www-form-urlencoded" {
		m.Params, err = url.ParseQuery(m.Body)
		if err != nil {
			return m, err
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
}

var tmplPackage = `
package {{.PackageName}}

type {{.StructName}} struct {
	
}

{{range .Methods}}
func ({{.StructNameFirstChar}} *{{.StructName}}) {{.MethodMame}}() (resp string, err error) {
	{{if .Params}}
	params := url.Values{}
	{{range $key, $value :=  .Params -}}
		params.Add("{{$key}}", "{{$value}}")
	{{end -}}
	{{end}}	

	for i := 0; i < .ReTryTimes; i++ {
		{{ if .ReqMethod }}
		req := hihttp.{{.ReqMethod}}("{{.URL}}")
		{{range $key, $value :=  .Heads -}}
		req.Header("{{$key}}", "{{$value}}")
		{{end -}}
		{{if .Params}}
		{{range $key, $value :=  .Params -}}
		req.Param("{{$key}}", "{{$value}}")
		{{end -}}
		{{end}}

		resp, err = req.String()
		if err == nil {
			break
		}

		buslog.GSLog.Error({{.StructNameFirstChar}}.LogPrefix+"{{.MethodMame}} 请求失败 resp:%s err:%s", resp, err.Error())

		utils.WaitRandMs(300, 500)
	}

	return 
}	
{{end}}
`
