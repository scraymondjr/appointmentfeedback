package main

import (
	"github.com/scraymondjr/appointment/cmd/cli/commander"
	"github.com/scraymondjr/appointment/datastore/neo4j"
)

func main() {
	store := neo4j.New()
	cmd := commander.Root(store)
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
