package commander

import (
	"github.com/spf13/cobra"

	"github.com/scraymondjr/appointment/internal"
)

func Root(store internal.Store) *cobra.Command {
	var root cobra.Command
	root.AddCommand(PatientCommand(store))
	return &root
}
