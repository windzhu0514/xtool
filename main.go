package main

import (
	"fmt"

	"github.com/windzhu0514/xtool/cmd"
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

	mycmd := cmd.NewCmd("xtool", &cmd.CmdPrompt{
		UsageLine: "xtool <command>  [arguments]",
		Short:     "short:xtool is a tool collect.",
		Long:      "long:xtool is a tool collect.",
	})

	fmt.Println(mycmd.Run())
}

// }

// 每个功能单独定义自己的flag解析和useage
