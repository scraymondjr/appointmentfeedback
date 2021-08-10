package internal

import (
	"encoding/json"
	"io"
	"reflect"

	"github.com/pkg/errors"
)

// ingest.go contains the API to parse and save a resource blob

// Ingest reads JSON data from reader and saves resources in the provided store.
//
// Returns an error if problem reading from reader, decoding JSON blob(s), or saving resource.
func Ingest(reader io.Reader, writer ResourceWriter) error {
	// parse data for resourceType

	decoder := json.NewDecoder(reader)

	resource, err := unmarshalResource(decoder)
	if err != nil {
		return err
	}

	// save resource

	return errors.Wrap(
		WriteResource(resource, writer),
		"problem saving resource "+resource.ID(),
	)
}

type ResourceWriter interface {
	WritePatient(Patient) error
	WriteDoctor(Doctor) error
	WriteAppointment(Appointment) error
	WriteDiagnosis(Diagnosis) error
}

func WriteResource(r Resource, writer ResourceWriter) error {
	switch r := r.(type) {
	case Bundle:
		// TODO guard against number of resources allowed to be written in one bundle
		// TODO include all bundled resources in one txn?
		for _, bundledResource := range r.Resources {
			if err := WriteResource(bundledResource, writer); err != nil {
				// TODO decide if should error-fast or continue to loop through all remaining resources
				return errors.Wrap(err, "problem writing resource "+bundledResource.ID())
			}
		}
	case Patient:
		return writer.WritePatient(r)
	case Doctor:
		return writer.WriteDoctor(r)
	case Appointment:
		return writer.WriteAppointment(r)
	case Diagnosis:
		return writer.WriteDiagnosis(r)
	default:
		return errors.Errorf("unknown resource type " + r.Type())
	}

	return nil
}

func unmarshalResource(decoder *json.Decoder) (Resource, error) {
	var m map[string]json.RawMessage
	if err := decoder.Decode(&m); err != nil {
		return nil, errors.Wrap(err, "problem decoding JSON object")
	}
	resourceTypeJSON, ok := m["resourceType"]
	if !ok {
		return nil, errors.Errorf("resourceType not found")
	}
	var resourceType string
	if err := json.Unmarshal(resourceTypeJSON, &resourceType); err != nil {
		return nil, errors.Wrap(err, "problem parsing resourceType as string")
	}
	resourceFactory, ok := resourceFactoriesByType[resourceType]
	if !ok {
		return nil, errors.Errorf("unknown resourceType: " + string(resourceType))
	}
	resource := resourceFactory()

	mJSON, _ := json.Marshal(m)
	if err := json.Unmarshal(mJSON, resource); err != nil {
		return nil, errors.Wrapf(err, "problem creating %s resource from input", resourceType)
	}

	return reflect.ValueOf(resource).Elem().Interface().(Resource), nil
}

var resourceFactoriesByType = map[string]func() Resource{
	"Bundle": func() Resource {
		return &Bundle{}
	},
	"Patient": func() Resource {
		return &Patient{}
	},
	"Doctor": func() Resource {
		return &Doctor{}
	},
	"Appointment": func() Resource {
		return &Appointment{}
	},
	"Diagnosis": func() Resource {
		return &Diagnosis{}
	},
}
