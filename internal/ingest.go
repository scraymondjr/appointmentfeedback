package internal

import (
	"encoding/json"
	"io"

	"github.com/pkg/errors"
)

// ingest.go contains the API to parse and save a resource blob

// Ingest reads JSON data from reader and saves resources in the provided store.
//
// Returns an error if problem reading from reader, problem
func Ingest(data io.Reader, s Store) error {
	// parse data for resourceType

	decoder := json.NewDecoder(data)

	resource, err := unmarshalResource(decoder)
	if err != nil {
		return err
	}

	// save resource

	return errors.Wrap(
		s.WriteResource(resource),
		"problem saving resource "+resource.ID(),
	)
}

