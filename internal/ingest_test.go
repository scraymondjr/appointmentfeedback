package internal_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/scraymondjr/appointment/datastore"
	. "github.com/scraymondjr/appointment/internal"
)

func TestIngest(t *testing.T) {
	f, err := os.Open("testdata/bundle.json")
	require.NoError(t, err)

	store := datastore.NewMemStore()
	err = Ingest(f, store)
	require.NoError(t, err)

	// soft-check that all data is saved in store

	assert.Contains(t, store.Patients, "6739ec3e-93bd-11eb-a8b3-0242ac130003")
	assert.Contains(t, store.Doctors, "9bf9e532-93bd-11eb-a8b3-0242ac130003")
	assert.Contains(t, store.Appointments, "be142dc6-93bd-11eb-a8b3-0242ac130003")
	assert.Contains(t, store.Diagnoses, "541a72a8-df75-4484-ac89-ac4923f03b81")
}

func TestReference_UnmarshalJSON(t *testing.T) {
	for name, tt := range map[string]struct {
		InputJSON      json.RawMessage
		ExpectedOutput Reference
		ExpectedError  error
	}{
		"happy path": {
			InputJSON: json.RawMessage(`{"reference": "Appointment/be142dc6-93bd-11eb-a8b3-0242ac130003"}`),
			ExpectedOutput: Reference{
				ResourceID:   "be142dc6-93bd-11eb-a8b3-0242ac130003",
				ResourceType: "Appointment",
			},
		},
		"bad reference format": {
			InputJSON:     json.RawMessage(`{"reference": "Appointment"}`),
			ExpectedError: fmt.Errorf(`unknown reference format for value "Appointment": expected 2 parts, but found 1`),
		},
		"not json object": {
			InputJSON:     json.RawMessage(`value`),
			ExpectedError: fmt.Errorf("invalid character 'v' looking for beginning of value"), // error from json parser
		},
		"no reference field": {
			InputJSON:     json.RawMessage(`{"another_field": "Appointment"}`),
			ExpectedError: fmt.Errorf(`unknown reference format for value "": expected 2 parts, but found 1`),
		},
	} {
		t.Run(name, func(t *testing.T) {
			var ref Reference
			err := json.Unmarshal(tt.InputJSON, &ref)
			if tt.ExpectedError != nil {
				assert.EqualError(t, err, tt.ExpectedError.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.ExpectedOutput, ref)
			}
		})
	}
}
