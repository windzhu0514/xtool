package saz2go

import (
	"archive/zip"
	"bufio"
	"errors"
	"io"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/PuerkitoBio/goquery"
)

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
		return errors.New("invalid fiddler saz file format")
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

func (s *saz2go) parseRequest(index int, rc io.Reader) (m oneMethod, err error) {
	m.Heads = make(map[string]string)
	m.Params = make(map[string]string)
	m.RetryTimes = 3
	m.StructName = "structName"
	m.StructNameFirstChar = "s"

	ss := bufio.NewScanner(rc)
	haveReadReqLine := false
	haveReadHeads := false
	isParamLine := false

	for ss.Scan() {
		line := ss.Text()
		if !haveReadReqLine {
			s1 := strings.Index(line, " ")
			s2 := strings.Index(line[s1+1:], " ")
			if s1 < 0 || s2 < 0 {
				log.Println("解析请求头失败")
				err = errors.New("解析请求头失败")
				return
			}
			s2 += s1 + 1
			m.ReqMethod = strings.Title(strings.ToLower(line[:s1]))

			URI := line[s1+1 : s2]

			URL, err2 := url.Parse(URI)
			if err2 != nil {
				err = err2
				return
			}

			m.URL = URL.Scheme + "://" + URL.Host + URL.Path

			// 设置函数名
			path := URL.Path
			path = strings.TrimSuffix(path, "/")
			index := strings.LastIndex(path, "/")
			if index < 0 {
				m.MethodMame = "defaultMethod" + strconv.Itoa(index)
			} else {
				lastStr := URL.Path[index+1:]
				dotIndex := strings.LastIndex(lastStr, ".")
				if dotIndex > 0 {
					lastStr = lastStr[:dotIndex]
				}

				if len(lastStr) == 0 {
					m.MethodMame = "defaultMethod" + strconv.Itoa(index)
				} else {
					m.MethodMame = lastStr
				}
			}

			// 解析参数
			params, err2 := url.ParseQuery(URL.RawQuery)
			if err2 != nil {
				err = err2

				log.Println("解析参数失败")
				haveReadReqLine = true
				continue
			}

			for k := range params {
				m.Params[k] = params.Get(k)
			}

			haveReadReqLine = true
			continue
		}

		if len(line) > 0 {
			if !isParamLine {
				headSlice := strings.Split(line, ": ")
				if headSlice[0] != "Cookie" && headSlice[0] != "Content-Length" {
					m.Heads[headSlice[0]] = headSlice[1]
				}

				haveReadHeads = true
			} else {
				// 根据Content-Type判断请求数据类型
				params, err2 := url.ParseQuery(line)
				if err2 != nil {
					err = err2

					log.Println("解析参数失败")
					continue
				}

				for k := range params {
					m.Params[k] = params.Get(k)
				}
			}
		} else {
			if haveReadHeads && haveReadReqLine { // 下一行是参数行
				isParamLine = true
				continue
			}
		}
	}

	return
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
	Heads               map[string]string
	Params              map[string]string
}

var tmplPackage = `
package {{.PackageName}}

type {{.StructName}} struct {
	
}

{{range .Methods}}
func ({{.StructNameFirstChar}} *{{.StructName}}) {{.MethodMame}}() (resp string, err error) {
	
	for i := 0; i < conf.GSystemConfig.ReTryTimes; i++ {
		req := httpclient.{{.ReqMethod}}("{{.URL}}")
		{{range $key, $value :=  .Heads -}}
		req.Header("{{$key}}", "{{$value}}")
		{{end -}}
		{{if .Params}}
		{{range $key, $value :=  .Params -}}
		req.Param("{{$key}}", "{{$value}}")
		{{end -}}
		{{end}}
		req.SetCookieJar({{.StructNameFirstChar}}.ci.CICookieJar)

		if {{.StructNameFirstChar}}.ci.UseProxy {
			req.SetAuthProxy({{.StructNameFirstChar}}.ci.ProxyUser, {{.StructNameFirstChar}}.ci.ProxyPass, {{.StructNameFirstChar}}.ci.ProxyIp, {{.StructNameFirstChar}}.ci.ProxyPort)
		}

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
