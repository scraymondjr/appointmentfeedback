package commander

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"

	"github.com/scraymondjr/appointment/internal"
)

func SurveyCommand(s internal.Store) *cobra.Command {
	return &cobra.Command{
		Use: "patient patient_id",
		Run: func(_ *cobra.Command, args []string) {
			patientID := args[0]
			runPatient(patientID, s)
		},
		Args: cobra.ExactArgs(1),
	}
}

type Prompt struct {
	PatientID string
	Store     internal.Store

	feedback *feedbackSurvey
}

type feedbackSurvey struct {
	Appointment *internal.Appointment
	Doctor      *internal.Doctor
	internal.Feedback
}

func (p *Prompt) startFeedback(appointment *internal.Appointment) {
	p.feedback = &feedbackSurvey{Appointment: appointment}
}

func (p *Prompt) patientDetails() {
	patient, err := p.Store.GetPatient(p.PatientID)
	if err != nil {
		fmt.Println("problem reading patient data: " + err.Error())
		return
	}

	json.NewEncoder(os.Stdin).Encode(patient)
}

func (p *Prompt) appointments() {
	appointments, err := p.Store.GetPatientAppointments(p.PatientID)
	if err != nil {
		fmt.Println("problem reading appoints for patient: " + err.Error())
		return
	}

	for _, appointment := range appointments {
		feedbackMsg := "Feedback survey available!"
		if appointment.Feedback != nil {
			feedbackMsg = "Feedback submitted"
		}
		fmt.Printf("appointment %s (%s) - %s\n", appointment.ID(), appointment.Status, feedbackMsg)
	}
}

func runPatient(patientID string, store internal.Store) {
	defer handleExit()
	p := &Prompt{PatientID: patientID, Store: store}
	prompt.New(
		func(in string) {
			in = strings.TrimSpace(in)
			var err error

			if p.feedback != nil {
				switch {
				case p.feedback.Recommend == 0:
					p.feedback.Recommend, err = strconv.Atoi(in)
					if err != nil {
						fmt.Printf("Sorry, your response %s was not understood. Please try again.\n", in)
						return
					}
					if p.feedback.Recommend < 1 || p.feedback.Recommend > 10 {
						fmt.Printf("Please enter a value between 1-10.\n")
						p.feedback.Recommend = 0
						return
					}

					fmt.Printf("\nThank you. You were diagnosed with %s. Did Dr %s explain how to manage this diagnosis in a way you could understand?\n\n", p.feedback.Appointment.Diagnosis.Name, p.feedback.Doctor.Name[0].Family)
				case p.feedback.Explained == nil:
					yesNo, _ := strconv.ParseBool(in)
					p.feedback.Explained = &yesNo

					fmt.Printf("\nWe appreciate the feedback, one last question: how do you feel about being diagnosed with %s?\n\n", p.feedback.Appointment.Diagnosis.Name)
				default:
					p.feedback.Feeling = &in

					// save
					// p.Store.SaveFeedback(p.feedback.Feedback)

					yesNo := "Yes"
					if !*p.feedback.Explained {
						yesNo = "No"
					}

					fmt.Printf("Thanks again! Here’s what we heard:\n\n")
					fmt.Printf("Your recommendation of Dr %s (1 - 10): %v\n", p.feedback.Doctor.Name[0].Family, p.feedback.Recommend)
					fmt.Printf("Dr %s explained your diagnosis of %s to you: %s\n", p.feedback.Doctor.Name[0].Family, p.feedback.Appointment.Diagnosis.Name, yesNo)
					fmt.Printf("Your feelings about your diagnosis: %s\n", *p.feedback.Feeling)

					p.feedback = nil
				}

				return
			}

			blocks := strings.Split(in, " ")
			switch blocks[0] {
			case "me":
				p.patientDetails()
			case "appointments":
				p.patientDetails()
			case "givefeedback":
				appointment, err := store.GetAppointment(blocks[1])
				if err != nil {
					fmt.Println("problem reading appoints for patient: " + err.Error())
					return
				}
				if appointment == nil {
					fmt.Printf("appointment %s not found for patient, cannot complete feedback\n", blocks[1])
					return
				}
				p.startFeedback(appointment)

				patient, _ := store.GetPatient(patientID)
				doctor, _ := store.GetDoctor(appointment.Actor.ResourceID)
				p.feedback.Doctor = doctor

				fmt.Printf("\nHi %s, on a scale of 1-10, would you recommend Dr %s to a friend or family member? 1 = Would not recommend, 10 = Would strongly recommend\n\n", patient.Name[0].Given[0], doctor.Name[0].Family)
			}
		},
		completer(p),
		prompt.OptionTitle("patient-feedback"),
		prompt.OptionPrefix("> "),
		prompt.OptionLivePrefix(func() (prefix string, useLivePrefix bool) {
			if p.feedback != nil {
				switch {
				case p.feedback.Recommend == 0:
					return fmt.Sprintf("(1 - 10): "), true
				case p.feedback.Explained == nil:
					return fmt.Sprintf("(Yes/No): "), true
				}
			}
			return "", false
		}),
	).Run()
}

func completer(p *Prompt) func(in prompt.Document) []prompt.Suggest {
	return func(in prompt.Document) []prompt.Suggest {
		if in.TextBeforeCursor() == "" || p.feedback != nil {
			return []prompt.Suggest{}
		}
		args := strings.Split(in.TextBeforeCursor(), " ")
		w := in.GetWordBeforeCursor()
		switch args[0] {
		case "givefeedback":
			appointments, err := p.Store.GetPatientAppointments(p.PatientID)
			if err != nil {
				fmt.Println("problem reading appoints for patient: " + err.Error())
				return []prompt.Suggest{}
			}

			prompts := make([]prompt.Suggest, len(appointments))
			for i := range appointments {
				if appointments[i].Feedback == nil {
					prompts[i] = prompt.Suggest{Text: appointments[i].ID()}
				}
			}
			return prompts
		}

		return prompt.FilterHasPrefix(suggestions, w, true)
	}
}

var suggestions = []prompt.Suggest{
	{"me", "Display info about me"},
	{"appointments", "List appointments"},
	{"givefeedback", "Provide feedback about a completed appointment"},
}

// hack to fix terminal prompt being disabled after exiting
// https://github.com/c-bata/go-prompt/issues/228#issuecomment-820639887
func handleExit() {
	rawModeOff := exec.Command("/bin/stty", "-raw", "echo")
	rawModeOff.Stdin = os.Stdin
	_ = rawModeOff.Run()
	rawModeOff.Wait()
}
