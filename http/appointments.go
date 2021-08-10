package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/scraymondjr/appointment/datastore"
	"github.com/scraymondjr/appointment/internal"
)

type appointmentsHandler struct {
	store datastore.Store
}

func (h appointmentsHandler) AddRoutes(g *echo.Group) {
	g.POST("/:appointmentId/feedback", h.POSTAppointmentFeedback)
	g.GET("/:appointmentId/feedback", h.GETAppointmentFeedback)
}

func (h appointmentsHandler) POSTAppointmentFeedback(c echo.Context) error {
	var feedbackRequest internal.Feedback
	if err := c.Bind(&feedbackRequest); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "could not parse request body").SetInternal(errors.WithStack(err))
	}

	err := h.store.SavePatientFeedback(c.Param("appointmentId"), feedbackRequest)
	if err != nil {
		return errors.Wrap(err, "problem saving feedback")
	}

	return c.NoContent(http.StatusCreated)
}

func (h appointmentsHandler) GETAppointmentFeedback(c echo.Context) error {
	appointmentID := c.Param("appointmentId")
	appointment, err := h.store.GetAppointment(appointmentID)
	if err != nil {
		return errors.Wrap(err, "problem getting appointment "+appointmentID)
	}
	if appointment == nil {
		return c.NoContent(http.StatusNotFound)
	}

	return c.JSON(http.StatusOK, appointment)
}
