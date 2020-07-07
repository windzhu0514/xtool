package saz2go

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	Cmd.Flags().StringVar(&ss.structName, "n", "strucName", "struct name")
	Cmd.Flags().StringVar(&ss.structFirstChar, "m", "", "method receiver name,default name first char of  struct name")
	Cmd.Flags().StringVar(&ss.outputFileName, "o", "", "output file name")
	Cmd.Flags().StringVar(&ss.tmplFileName, "t", "", "template file name")
}

var Cmd = &cobra.Command{
	Use:   "saz2go",
	Short: "saz2go convert fiddler .saz file to go file",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("specify file to convert")
			return
		}

		if err := ss.Convert(args[0]); err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("generate success :)")
	},
}

var CmdParse = &cobra.Command{
	Use:   "sazparse",
	Short: "sazparse parse fiddler .saz file",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("specify file to convert")
			return
		}

		if err := ss.Parse(args[0]); err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("generate success :)")
	},
}
