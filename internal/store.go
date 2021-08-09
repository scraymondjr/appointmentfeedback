package internal

import "github.com/pkg/errors"

func NewStore() Store {
	return Store{
		Patients:     map[string]Patient{},
		Doctors:      map[string]Doctor{},
		Appointments: map[string]Appointment{},
		Diagnoses:    map[string]Diagnosis{},
	}
}

type Store struct {
	Patients     map[string]Patient
	Doctors      map[string]Doctor
	Appointments map[string]Appointment
	Diagnoses    map[string]Diagnosis
}

// patient -> appointment -> survey
//
// If patient -> appointment.IsComplete && appointment -> survey == nil, then prompt for survey

func (s Store) WriteResource(r Resource) error {
	switch r := r.(type) {
	case Bundle:
		// TODO guard against number of resources allowed to be written in one bundle
		for _, bundledResource := range r.Resources {
			if err := s.WriteResource(bundledResource); err != nil {
				// TODO decide if should error-fast or continue to loop through all remaining resources
				return errors.Wrap(err, "problem writing resource "+bundledResource.ID())
			}
		}
	case Patient:
		s.Patients[r.ID()] = r
	case Doctor:
		s.Doctors[r.ID()] = r
	case Appointment:
		s.Appointments[r.ID()] = r
	case Diagnosis:
		s.Diagnoses[r.ID()] = r
	default:
		return errors.Errorf("unknown resource type " + r.Type())
	}

	return nil
}
