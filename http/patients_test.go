package http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/scraymondjr/appointment/datastore"
	"github.com/scraymondjr/appointment/internal"
)

func TestPatientsHandler_GETPatientAppointments(t *testing.T) {
	const patientID = "testpatient"
	appointments := []internal.Appointment{
		{Status: "finished"}, {Status: "other"},
	}
	store := fakeStore{
		patientAppointments: map[string][]internal.Appointment{
			patientID: appointments,
		},
	}

	handler := patientsHandler{store}

	e := echo.New()
	handler.AddRoutes(e.Group(""))

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s/appointments", patientID), nil)
	resp := httptest.NewRecorder()
	e.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)

	// TODO fix Appointment marshalling to be able to unmarshal JSON back into Go type
	// var response []internal.Appointment
	// err := json.NewDecoder(resp.Body).Decode(&response)
	// require.NoError(t, err)
	//
	// assert.ElementsMatch(t, appointments, response)
}

type fakeStore struct {
	datastore.Store

	patientAppointments map[string][]internal.Appointment
}

func (f fakeStore) GetPatientAppointments(patientID string) ([]internal.Appointment, error) {
	return f.patientAppointments[patientID], nil
}
