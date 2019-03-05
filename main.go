package main

import (
	"fmt"

	"github.com/windzhu0514/cmd"
	"github.com/windzhu0514/xtool/saz2go"
)

func main() {
	// flag.Parse()
	// flag.Usage = func() {
	// 	fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
	// 	flag.PrintDefaults()
	// }
	//
	// if flag.NArg() < 1 {
	// 	flag.Usage()
	// 	return
	// }
	//
	// switch flag.Arg(0) {
	// case "saz2go":
	// default:
	// 	flag.Usage()
	// 	return
	// }

	cmd.NewBaseCmd("xtool", &cmd.CmdPrompt{
		UsageLine: "xtool <command>  [arguments]",
		Short:     "xtool is a tool collect",
		Long:      "xtool is a tool collect",
	}).Exe = saz2go.New()

	// baseCmd.AddSubCmd("saz2go", &cmd.CmdPrompt{
	// 	UsageLine: "saz2go [arguments]",
	// 	Short:     "transform fiddler sessions to go code",
	// 	Long:      "saz2go is a tool transform fiddler sessions to go code",
	// }).Exe = saz2go.New()

	if err := cmd.Run(); err != nil {
		fmt.Println(err)
	}
}
