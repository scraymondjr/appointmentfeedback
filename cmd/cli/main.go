package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/scraymondjr/appointment/internal"
	"github.com/scraymondjr/appointment/internal/commander"
)

func init() {
	root.AddCommand(ingest)
}

func main() {
	store := internal.NewNeo4jStore()
	cmd := commander.Root(store)
	cmd.Execute()
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
