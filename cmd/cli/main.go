package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	root.AddCommand(ingest)
}

func main() {
	root.Execute()
}

var (
	root = &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
	ingest = &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("ingest file" + args[0])
		},
		Use:  "ingest input_file.json",
		Args: cobra.ExactArgs(1),
	}
)
