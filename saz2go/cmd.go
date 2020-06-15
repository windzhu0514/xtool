package saz2go

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	Cmd.Flags().StringVar(&ss.structName, "n", "strucName", "struct name")
	Cmd.Flags().StringVar(&ss.structFirstChar, "m", "r", "method receiver name")
	Cmd.Flags().StringVar(&ss.outputFileName, "o", "", "output file name")
	Cmd.Flags().StringVar(&ss.tmplFileName, "t", "", "template file name")
}

var Cmd = &cobra.Command{
	Use:   "saz2go",
	Short: "saz2go conversion fiddler .saz file to go file",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if err := ss.Run(args[0]); err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("generate success :)")
	},
}
