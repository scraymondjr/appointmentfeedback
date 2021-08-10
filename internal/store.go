package internal

type Store interface {
	GetPatient(id string) (*Patient, error)
	GetDoctor(id string) (*Doctor, error)
	GetPatientAppointments(patientID string) ([]Appointment, error)
	SavePatientFeedback(appointmentID string, feedback Feedback) error
	GetPatientFeedback(appointmentID string) (*Feedback, error)
	GetAppointment(id string) (*Appointment, error)
}

func NewMemStore() MemStore {
	return MemStore{
		Patients:     map[string]Patient{},
		Doctors:      map[string]Doctor{},
		Appointments: map[string]Appointment{},
		Diagnoses:    map[string]Diagnosis{},
	}
}

type MemStore struct {
	Store
	Patients     map[string]Patient
	Doctors      map[string]Doctor
	Appointments map[string]Appointment
	Diagnoses    map[string]Diagnosis
}

func (s MemStore) WritePatient(patient Patient) error {
	s.Patients[patient.ID()] = patient
	return nil
}

func (s MemStore) WriteDoctor(doctor Doctor) error {
	s.Doctors[doctor.ID()] = doctor
	return nil
}

func (s MemStore) WriteAppointment(appointment Appointment) error {
	s.Appointments[appointment.ID()] = appointment
	return nil
}

func (s MemStore) WriteDiagnosis(diagnosis Diagnosis) error {
	s.Diagnoses[diagnosis.ID()] = diagnosis
	return nil
}
