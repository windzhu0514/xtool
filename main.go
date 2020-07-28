package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/windzhu0514/xtool/config"
	"github.com/windzhu0514/xtool/saz2go"
)

func main() {
	if err := config.LoadConfig(); err != nil {
		fmt.Println(err)
		return
	}

	f, err := os.OpenFile(config.LogFilename(), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(f)

	var rootCmd = &cobra.Command{
		Use:  "xtool",
		Long: `xtool is a tool collect`,
	}

	rootCmd.AddCommand(saz2go.Cmd)
	rootCmd.AddCommand(saz2go.CmdParse)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		return
	}
}
