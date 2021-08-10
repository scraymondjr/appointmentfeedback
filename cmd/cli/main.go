package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/scraymondjr/appointment/cmd/cli/commander"
	"github.com/scraymondjr/appointment/datastore/neo4j"
)

func init() {
	root.AddCommand(ingest)
}

func main() {
	store := neo4j.New()
	cmd := commander.Root(store)
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
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
