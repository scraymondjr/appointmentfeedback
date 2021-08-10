package http

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/scraymondjr/appointment/datastore"
)

// Echo returns an echo.Echo instance configured with all handlers.
func Echo(store datastore.Store) *echo.Echo {
	e := echo.New()
	e.Use(
		middleware.Logger(),
		middleware.Recover(),
	)

	patientsHandler{store: store}.AddRoutes(e.Group("/patients", func(next echo.HandlerFunc) echo.HandlerFunc {
		// check if authorized to access patients resource
		return next
	}))
	appointmentsHandler{store: store}.AddRoutes(e.Group("/appointments", func(next echo.HandlerFunc) echo.HandlerFunc {
		// check if authorized to access appointments resource
		return next
	}))

	return e
}
