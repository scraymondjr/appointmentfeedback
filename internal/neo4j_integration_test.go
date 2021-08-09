// +build integration

package internal

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNeo4jStoreIngest(t *testing.T) {
	f, err := os.Open("testdata/bundle.json")
	require.NoError(t, err)

	store := NewNeo4jStore()
	err = Ingest(f, store)
	require.NoError(t, err)

	// soft-check that all data is saved in store

	patient, err := store.GetPatient("6739ec3e-93bd-11eb-a8b3-0242ac130003")
	require.NoError(t, err)
	require.NotNil(t, patient)

	appointments, err := store.GetPatientAppointments(patient.ResourceID)
	require.NoError(t, err)
	assert.NotEmpty(t, appointments)

	appointment, err := store.GetAppointment("be142dc6-93bd-11eb-a8b3-0242ac130003")
	require.NoError(t, err)
	require.NotNil(t, appointment)
	assert.Equal(t, "Diabetes without complications", appointment.Diagnosis.Name)
}

func TestNeo4jStore_WriteResource(t *testing.T) {
	store := NewNeo4jStore()

	appointmentJSON := `{
        "resourceType": "Appointment",
        "id": "be142dc6-93bd-11eb-a8b3-0242ac130002",
        "status": "finished",
        "type": [
          {
            "text": "Endocrinologist visit"
          }
        ],
        "subject": {
          "reference": "Patient/6739ec3e-93bd-11eb-a8b3-0242ac130003"
        },
        "actor": {
          "reference": "Doctor/9bf9e532-93bd-11eb-a8b3-0242ac130003"
        },
        "period": {
          "start": "2021-04-02T11:30:00Z",
          "end": "2021-04-02T12:00:00Z"
        }
      }`

	var appointment Appointment
	err := json.Unmarshal([]byte(appointmentJSON), &appointment)
	require.NoError(t, err)

	err = store.WriteResource(appointment)
	require.NoError(t, err)

	storedAppointment, err := store.GetAppointment(appointment.ID())
	require.NoError(t, err)
	assert.Equal(t, &appointment, storedAppointment)

	storedPatient, err := store.GetPatient("6739ec3e-93bd-11eb-a8b3-0242ac130003")
	require.NoError(t, err)
	assert.Equal(t, &Patient{
		ResourceTypeAndID: ResourceTypeAndID{
			ResourceID:   "6739ec3e-93bd-11eb-a8b3-0242ac130003",
			ResourceType: "Patient",
		},
	}, storedPatient)
}
