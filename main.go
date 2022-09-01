package main

import (
	"github.com/nodejayes/pgapp_cli/cmd"
	"github.com/spf13/cobra"
	"log"
)

func main() {
	var rootCmd = &cobra.Command{Use: "app"}
	rootCmd.AddCommand(cmd.Create)
	log.Fatal(rootCmd.Execute())
}
