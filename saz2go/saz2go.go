package saz2go

import (
	"archive/zip"
	"bufio"
	"errors"
	"flag"
	"fmt"
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
	tmplFileName    string
	log             *log.Logger
}

var ss saz2go

func New() *saz2go {
	var s saz2go

	s.log = log.New(os.Stderr, "saz2go", log.Ldate|log.Ltime|log.Lshortfile)

	return &s
}

func (s *saz2go) Bind(flagSet *flag.FlagSet) {
	flagSet.StringVar(&s.structName, "sn", "strucName", "specified struct name")
	flagSet.StringVar(&s.structFirstChar, "sfc", "s", "specified struct first char name")
	flagSet.StringVar(&s.tmplFileName, "tf", "", "specified template file name")
}

func (s *saz2go) Run(args []string) error {
	if len(args) < 1 {
		return errors.New("please specify a file name ")
	}

	fileName := args[0]

	r, err := zip.OpenReader(fileName)
	if err != nil {
		return err
	}
	defer r.Close()

	var pack onePackage
	pack.PackageName = "packagename"
	pack.StructName = "structname"
	pack.StructNameFirstChar = "s"

	if s.structName != "" {
		pack.StructName = s.structName
	}

	if s.structFirstChar != "" {
		pack.StructNameFirstChar = s.structFirstChar
	}

	files := make(map[string]*zip.File)
	for _, f := range r.File {
		files[f.Name] = f // raw/02_s.txt
	}

	indexFile, exist := files["_index.htm"]
	if !exist {
		s.log.Println("文件内容错误")
		return errors.New("文件内容错误")
	}

	read, err := indexFile.Open()
	if err != nil {
		s.log.Println(err)
		return err
	}

	doc, err := goquery.NewDocumentFromReader(read)
	if err != nil {
		s.log.Println(err)
		return err
	}

	//jsonStrs := [][]byte{}
	doc.Find("body table tbody tr").Each(func(i int, selection *goquery.Selection) {

		reqName, ok0 := selection.Find("td a").Eq(0).Attr("href")
		respName, ok1 := selection.Find("td a").Eq(1).Attr("href")

		reqName = strings.Replace(reqName, "\\", "/", -1)
		respName = strings.Replace(respName, "\\", "/", -1)
		//s.log.Println(reqName, respName)

		if ok0 && ok1 {
			reqFile, exist := files[reqName]
			if exist {
				reqRead, err := reqFile.Open()
				if err != nil {
					s.log.Println(err)
					return
				}

				if method, err := parseRequest(i, bufio.NewReader(reqRead)); err != nil {
					s.log.Println(err)
				} else {
					method.StructNameFirstChar = pack.StructNameFirstChar
					method.StructName = pack.StructName
					pack.Methods = append(pack.Methods, method)
				}

				reqRead.Close()
			}
		}
	})

	var t *template.Template
	if s.tmplFileName != "" {
		t, err = template.ParseFiles(s.tmplFileName)
	} else {
		t = template.New("req")
		t, err = t.Parse(tmplPackage)
	}

	if err != nil {
		s.log.Println(err)
		return err
	}

	//f, _ := os.Create("gen/gen.go")
	goFileName := fileName
	goFileName = strings.Replace(goFileName, ".saz", ".go", -1)
	f, err := os.OpenFile(goFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		s.log.Println(err)
		return err
	}
	if err := t.Execute(f, pack); err != nil {
		s.log.Println(err)
		return err
	}

	f.Close()

	fmt.Println("生成成功 :)")

	return nil
}

func parseRequest(count int, rc io.Reader) (m oneMethod, err error) {
	m.Heads = make(map[string]string)
	m.Params = make(map[string]string)
	m.RetryTimes = 3
	m.StructName = "structName"
	m.StructNameFirstChar = "s"

	s := bufio.NewScanner(rc)
	haveReadReqLine := false
	haveReadHeads := false
	isParamLine := false

	for s.Scan() {
		line := s.Text()
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
				m.MethodMame = "defaultMethod" + strconv.Itoa(count)
			} else {
				lastStr := URL.Path[index+1:]
				if len(lastStr) == 0 {
					m.MethodMame = "defaultMethod" + strconv.Itoa(count)
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

		buslog.GSLog.Error({{.StructNameFirstChar}}.LogPrefix+"{{.MethodMame}}请求失败 resp:%s err:%s", resp, err.Error())

		utils.WaitRandMs(300, 500)
	}

	return 
}	
{{end}}
`
