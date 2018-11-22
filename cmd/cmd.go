package cmd

import (
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"
)

var usageTemplate = `{{.Prompt.Long | trim}}

Usage:
	{{.Prompt.UsageLine}}

The commands are:
{{range .Commands}}{{if or (.Runnable) .Commands}}
	{{.Name | printf "%-11s"}} {{.Short}}{{end}}{{end}}

Use "go help{{with .Name}} {{.}}{{end}} <command>" for more information about a command.
`

func capitalize(s string) string {
	if s == "" {
		return s
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToTitle(r)) + s[n:]
}

func tmpl(w io.Writer, text string, data interface{}) {
	t := template.New("top")
	t.Funcs(template.FuncMap{"trim": strings.TrimSpace, "capitalize": capitalize})
	template.Must(t.Parse(text))

	err := t.Execute(w, data)
	if err != nil {
		panic(err)
	}
}

type executor interface {
	Bind()
	Run() error
}

type CmdPrompt struct {
	// UsageLine is the one-line usage message.
	// The first word in the line is taken to be the command name.
	UsageLine string

	// Short is the short description shown in the 'go help' output.
	Short string

	// Long is the long message shown in the 'go help <this-command>' output.
	Long string
}

type command struct {

	// Run runs the command.
	// The args are the arguments after the command name.
	//Run func(cmd *Command, args []string)
	//Run func(args []string) error
	exe executor

	Name string

	Prompt *CmdPrompt

	// Flag is a set of flags specific to this command.
	FlagSet flag.FlagSet

	// CustomFlags indicates that the command will do its own
	// flag parsing.
	CustomFlags bool

	// Commands lists the available commands and help topics.
	// The order here is the order in which they are printed by 'go help'.
	// Note that subcommands are in general best avoided.
	Commands []*command
}

var baseCmd command

func NewCmd(name string, prompt *CmdPrompt) *command {
	c := &command{
		Name:   name,
		Prompt: prompt,
	}

	return c
}

func (c *command) Add(cmd *command) {

	for _, v := range c.Commands {
		if v.Name == cmd.Name {
			return
		}
	}

	c.Commands = append(c.Commands, cmd)

	return
}

func (c *command) Usage() {
	fmt.Fprintln(os.Stderr, c.Prompt.Short)
	//fmt.Fprintf(os.Stderr, "usage: %s\n", c.Prompt.UsageLine)
	//fmt.Fprintf(os.Stderr, "Run 'go help %s' for details.\n", c.LongName())

	tmpl(os.Stderr, usageTemplate, c)
	os.Exit(2)
}

// type Command struct {
// Run:       runBug,
// UsageLine: "go bug",
// Short:     "start a bug report",
// Long: `
// Bug opens the default browser and starts a new bug report.
// The report includes useful system information.
// 	`,

type exe struct {
}

func (a *exe) Bind() {
}

func (a *exe) Run() error {
	return nil
}

var defaultExe = &exe{}

// func Add(cmd *command) {
//
// 	for _, v := range baseCmd.Commands {
// 		if v.Name == cmd.Name {
// 			return
// 		}
// 	}
//
// 	if cmd.exe == nil {
// 		cmd.exe = defaultExe
// 	}
//
// 	cmd.exe.Bind()
//
// 	baseCmd.Commands = append(baseCmd.Commands, cmd)
//
// 	return
// }

func usage() {
	//fmt.Fprintln(os.Stderr, baseCmd.Prompt.Short)
	//fmt.Fprintf(os.Stderr, "usage: %s\n", baseCmd.Prompt.UsageLine)
	//mt.Fprintf(os.Stderr, "Run 'go help %s' for details.\n", c.LongName())

	tmpl(os.Stderr, usageTemplate, c)
}

func Run() error {
	args := os.Args[1:]
	fmt.Println(args)
	if len(args) < 1 {
		usage()
		return nil
	}

	cmdName := args[0]
	fmt.Println(cmdName)
	for _, cmd := range baseCmd.Commands {
		if cmd.Name == cmdName {
			fmt.Println(args[1:])
			fmt.Println(cmd.FlagSet.Parse(args[1:]))
			cmd.exe.Run()
			return nil
		}
	}

	flag.Usage()
	return errors.New("args invalid")
}
