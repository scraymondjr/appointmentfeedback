package commander

import (
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/scraymondjr/appointment/internal"
)

func IngestCommand(writer internal.ResourceWriter) *cobra.Command {
	return &cobra.Command{
		Use: "ingest json_file_path",
		RunE: func(_ *cobra.Command, args []string) error {
			f, err := os.Open(args[0])
			if err != nil {
				return errors.Wrap(err, "problem opening file at "+args[0])
			}
			defer f.Close()
			if err := internal.Ingest(f, writer); err != nil {
				return errors.Wrap(err, "problem ingesting file "+args[0])
			}

			return nil
		},
		Args: cobra.ExactArgs(1),
	}
}
