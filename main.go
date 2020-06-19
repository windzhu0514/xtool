package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/windzhu0514/xtool/saz2go"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:  "xtool",
		Long: `xtool is a tool collect`,
	}

	rootCmd.AddCommand(saz2go.Cmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
