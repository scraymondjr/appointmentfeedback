package commander

import (
	"github.com/spf13/cobra"

	"github.com/scraymondjr/appointment/internal"
)

func SurveyCommand(s internal.Store) *cobra.Command {
	return &cobra.Command{
		Use: "patient patient_id appointments|surveys",
		Run: func(_ *cobra.Command, args []string) {
			patientID := args[0]
			switch args[1] {
			case "appointments":
				getAppointmentsForUser(patientID, s)
			}
		},
		Args: cobra.ExactArgs(2),
	}
}

func getAppointmentsForUser(id string, s internal.Store) []internal.Appointment {
	for _, appointment := range s.Appointments {
		if appointment.
	}
}
