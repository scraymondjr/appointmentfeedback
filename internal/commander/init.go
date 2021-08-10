package commander

import (
	"github.com/spf13/cobra"

	"github.com/scraymondjr/appointment/internal"
)

func Root(store interface {
	internal.Store
	internal.ResourceWriter
}) *cobra.Command {
	var root cobra.Command
	root.AddCommand(
		PatientCommand(store),
		IngestCommand(store),
	)
	return &root
}
