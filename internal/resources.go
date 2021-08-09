package internal

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type Resource interface {
	Type() string
	ID() string
}

type (
	Bundle struct {
		ResourceTypeAndID
		Resources BundledResources `json:"entry"`
	}
	BundledResources []Resource

	Patient struct {
		ResourceTypeAndID
		Name []Name `json:"name"` // just take first
		// TODO all patient fields
	}

	Name struct {
		Text   string   `json:"text"`
		Family string   `json:"family"`
		Given  []string `json:"given"` // just take first
	}

	Doctor struct {
		ResourceTypeAndID
		Name []Name `json:"name"` // just take first
	}

	Appointment struct {
		ResourceTypeAndID
		Status      string     `json:"status"`
		Description string     `json:"-"` // TODO
		Subject     Reference  `json:"subject"`
		Actor       Reference  `json:"actor"`
		Feedback    *Reference `json:"feedback"`
		Diagnosis   Diagnosis  `json:"-"`
	}

	Diagnosis struct {
		ResourceTypeAndID
		Status      string    `json:"status"`
		Name        string    `json:"code"` // TODO
		Appointment Reference `json:"appointment"`
	}

	Feedback struct {
		Recommend int     `json:"recommend"`
		Explained *bool   `json:"explained"`
		Feeling   *string `json:"feeling"`
	}

	ResourceTypeAndID struct {
		ResourceID   string `json:"id"`
		ResourceType string `json:"resourceType"`
	}

	Reference ResourceTypeAndID
)

func (r ResourceTypeAndID) Type() string {
	return r.ResourceType
}

func (r ResourceTypeAndID) ID() string {
	return r.ResourceID
}

func (bundled *BundledResources) UnmarshalJSON(data []byte) error {
	var entries []map[string]json.RawMessage
	if err := json.Unmarshal(data, &entries); err != nil {
		return errors.Wrap(err, "problem unmarshalling list of json objects")
	}

	resources := make([]Resource, len(entries))
	for i, entry := range entries {
		resource, err := unmarshalResource(json.NewDecoder(bytes.NewReader(entry["resource"])))
		if err != nil {
			return errors.Wrap(err, "problem unmarshalling bundled resource at entry "+strconv.Itoa(i))
		}
		resources[i] = resource
	}
	*bundled = resources
	return nil
}

// UnmarshalJSON unmarshals the code from the json data and sets that as the name.
func (n *Diagnosis) UnmarshalJSON(data []byte) error {
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	// set name to first coding found in diagnosis
	for _, coding := range m["code"].(map[string]interface{})["coding"].([]interface{}) {
		n.Name = coding.(map[string]interface{})["name"].(string)
		break
	}

	delete(m, "code")

	data, _ = json.Marshal(m)

	type alias Diagnosis
	if err := json.Unmarshal(data, (*alias)(n)); err != nil {
		return err
	}

	return nil
}

func (r *Reference) UnmarshalJSON(data []byte) error {
	var ref struct {
		Reference string `json:"reference"`
	}
	if err := json.Unmarshal(data, &ref); err != nil {
		return err
	}

	// expect reference format of "{ResourceType}/{ResourceID}"
	parts := strings.SplitN(ref.Reference, "/", 2)
	if len(parts) != 2 {
		return errors.Errorf(`unknown reference format for value "%s": expected %d parts, but found %d`, ref.Reference, 2, len(parts))
	}

	// TODO validate parts

	r.ResourceType = parts[0]
	r.ResourceID = parts[1]

	return nil
}
