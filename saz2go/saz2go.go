package saz2go

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/windzhu0514/xtool/cmd"
)

type saz2go struct {
	cmd cmd.Command

	structName string
	log        *log.Logger
}

func New() *saz2go {
	var s saz2go

	s.cmd.Name = "saz2go"
	s.cmd.Short = "转换fiddler为go代码"
	s.cmd.Long = "转换fiddler为go代码"
	s.cmd.FlagSet.StringVar(&s.structName, "s", "strucName", "specified struct name")
	s.log = log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)

	//s.cmd.Run = s.Run

	return &s
}

func (s *saz2go) Cmd() cmd.Command {
	return s.cmd
}

func (s *saz2go) Run(args []string) error {

	fmt.Println("参数", args)
	fmt.Println("结构体名", s.structName)

	fmt.Println("呜哈哈哈哈哈")

	return nil
	// r, err := zip.OpenReader(flag.Arg(0))
	// if err != nil {
	// 	s.log.Println(err)
	// 	return
	// }
	// defer r.Close()
	//
	// var pack onePackage
	// pack.PackageName = "test"
	// pack.StructName = "structName"
	//
	// files := make(map[string]*zip.File)
	// for _, f := range r.File {
	// 	files[f.Name] = f // raw/02_s.txt
	// }
	//
	// indexFile, exist := files["_index.htm"]
	// if !exist {
	// 	s.log.Println("文件内容错误")
	// 	return
	// }
	//
	// read, err := indexFile.Open()
	// if err != nil {
	// 	s.log.Println(err)
	// 	return
	// }
	//
	// doc, err := goquery.NewDocumentFromReader(read)
	// if err != nil {
	// 	s.log.Println(err)
	// 	return
	// }
	//
	// //jsonStrs := [][]byte{}
	// doc.Find("body table tbody tr").Each(func(i int, selection *goquery.Selection) {
	//
	// 	reqName, ok0 := selection.Find("td a").Eq(0).Attr("href")
	// 	respName, ok1 := selection.Find("td a").Eq(1).Attr("href")
	//
	// 	reqName = strings.Replace(reqName, "\\", "/", -1)
	// 	respName = strings.Replace(respName, "\\", "/", -1)
	// 	//s.log.Println(reqName, respName)
	//
	// 	if ok0 && ok1 {
	// 		reqFile, exist := files[reqName]
	// 		if exist {
	// 			reqRead, err := reqFile.Open()
	// 			if err != nil {
	// 				s.log.Println(err)
	// 				return
	// 			}
	//
	// 			if method, err := parseRequest(i, bufio.NewReader(reqRead)); err != nil {
	// 				s.log.Println(err)
	// 			} else {
	// 				pack.Methods = append(pack.Methods, method)
	// 			}
	//
	// 			reqRead.Close()
	// 		}
	// 	}
	// })
	//
	// t := template.New("req")
	// t, err = t.Parse(tmplPackage)
	// if err != nil {
	// 	s.log.Println(err)
	// 	return
	// }
	//
	// if err := os.Mkdir("gen", 0755); err != nil {
	// 	if !os.IsExist(err) {
	// 		s.log.Println(err)
	// 		return
	// 	}
	// }
	//
	// //f, _ := os.Create("gen/gen.go")
	// f, err := os.OpenFile("gen/gen.go", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	// if err != nil {
	// 	s.log.Println(err)
	// 	return
	// }
	// if err := t.Execute(f, pack); err != nil {
	// 	s.log.Println(err)
	// 	return
	// }
	//
	// f.Close()
	//
	// fmt.Println("生成成功 :)")
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
			index := strings.LastIndex(URL.Path, "/")
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
	PackageName string
	StructName  string
	Methods     []oneMethod
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
