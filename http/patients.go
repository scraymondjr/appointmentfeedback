package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/scraymondjr/appointment/datastore"
)

type patientsHandler struct {
	store datastore.Store
}

func (h patientsHandler) AddRoutes(e *echo.Group) {
	e.GET("/:patientId/appointments", h.GETPatientAppointments, func(next echo.HandlerFunc) echo.HandlerFunc {
		// authorize request to access appointments for patient
		return next
	})
}

func (h patientsHandler) GETPatientAppointments(c echo.Context) error {
	patientID := c.Param("patientId")
	appointments, err := h.store.GetPatientAppointments(patientID)
	if err != nil {
		return errors.Wrap(err, "problem accessing patient appointments")
	}

	// TODO translate response into well-defined schema
	return c.JSON(http.StatusOK, appointments)
}
