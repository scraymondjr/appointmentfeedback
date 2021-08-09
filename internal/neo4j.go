package internal

import (
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/pkg/errors"
)

func NewNeo4jStore() Neo4jStore {
	// TODO accept config input instead of hardcoding
	driver, err := neo4j.NewDriver("neo4j://localhost:7687", neo4j.AuthToken{})
	if err != nil {
		panic(err)
	}
	return Neo4jStore{
		neo4j: driver,
	}
}

type Neo4jStore struct {
	neo4j neo4j.Driver
}

func (store Neo4jStore) WritePatient(p Patient) error {
	sess := store.neo4j.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer sess.Close()
	_, err := sess.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		return tx.Run(
			`MERGE (a:Patient {
					id: $id,
					givenName: $givenName,
					familyName: $familyName
				} ) RETURN a`,
			map[string]interface{}{
				"id":         p.ID(),
				"givenName":  p.Name[0].Given[0],
				"familyName": p.Name[0].Family,
			},
		)
	})
	if err != nil {
		return errors.Wrap(err, "problem saving patient "+p.ID())
	}
	return nil
}

func (store Neo4jStore) GetPatient(id string) (*Patient, error) {
	sess := store.neo4j.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer sess.Close()

	record, err := sess.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		result, err := tx.Run(`
		MATCH (p:Patient { id:$id })
		RETURN p
		`, map[string]interface{}{
			"id": id,
		})
		if err != nil {
			return nil, err
		}

		return result.Single()
	})
	if record == nil || err != nil {
		return nil, err
	}

	patientNode := record.(*neo4j.Record).Values[0].(neo4j.Node)

	return &Patient{
		ResourceTypeAndID: ResourceTypeAndID{
			ResourceID:   id,
			ResourceType: "Patient",
		},
		Name: []Name{
			{
				Family: patientNode.Props["familyName"].(string),
				Given:  []string{patientNode.Props["givenName"].(string)},
			},
		},
	}, nil
}

func (store Neo4jStore) WriteDoctor(d Doctor) error {
	sess := store.neo4j.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer sess.Close()
	_, err := sess.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		return tx.Run(
			`MERGE (a:Doctor {
					id: $id,
					givenName: $givenName,
					familyName: $familyName
				} ) RETURN a`,
			map[string]interface{}{
				"id":         d.ID(),
				"givenName":  d.Name[0].Given[0],
				"familyName": d.Name[0].Family,
			})
	})
	if err != nil {
		return errors.Wrap(err, "problem saving patient "+d.ID())
	}
	return nil
}

func (store Neo4jStore) GetDoctor(id string) (*Doctor, error) {
	sess := store.neo4j.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer sess.Close()

	record, err := sess.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		result, err := tx.Run(`
		MATCH (d:Doctor { id:$id })
		RETURN d
		`, map[string]interface{}{
			"id": id,
		})
		if err != nil {
			return nil, err
		}

		return result.Single()
	})
	if record == nil || err != nil {
		return nil, err
	}

	patientNode := record.(*neo4j.Record).Values[0].(neo4j.Node)

	return &Doctor{
		ResourceTypeAndID: ResourceTypeAndID{
			ResourceID:   id,
			ResourceType: "Doctor",
		},
		Name: []Name{
			{
				Family: patientNode.Props["familyName"].(string),
				Given:  []string{patientNode.Props["givenName"].(string)},
			},
		},
	}, nil
}

func (store Neo4jStore) GetPatientAppointments(patientID string) ([]Appointment, error) {
	sess := store.neo4j.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer sess.Close()

	result, err := sess.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		result, err := tx.Run(`
		MATCH (:Patient { id:$patientId })<-[:SUBJECT]-(a:Appointment)
		MATCH (a)-[r]-(n)
		RETURN *
		`, map[string]interface{}{
			"patientId": patientID,
		})
		if err != nil {
			return nil, err
		}

		return result.Collect()
	})
	if result == nil || err != nil {
		return nil, err
	}

	m := map[string]*Appointment{}
	for _, record := range result.([]*neo4j.Record) {
		processAppointmentRecord(record, m)
	}

	var apps []Appointment
	for _, app := range m {
		apps = append(apps, *app)
	}

	return apps, nil
}

func (store Neo4jStore) GetPatientNotifications(patientID string) error {
	return nil
}

func (store Neo4jStore) WriteAppointment(a Appointment) error {
	sess := store.neo4j.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer sess.Close()
	_, err := sess.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		return tx.Run(
			`MERGE (a:Appointment {
					id: $id,
					status: $status,
					type: $type
				} )
				MERGE (p:Patient { id:$patientId })
				MERGE (d:Doctor { id:$doctorId })
				MERGE (a)-[sub:SUBJECT]->(p)
				MERGE (a)-[actor:ACTOR]->(d)
				RETURN a`,
			map[string]interface{}{
				"id":        a.ID(),
				"status":    a.Status,
				"type":      a.Description,
				"patientId": a.Subject.ResourceID,
				"doctorId":  a.Actor.ResourceID,
			},
		)
	})
	if err != nil {
		return errors.Wrap(err, "problem saving patient "+a.ID())
	}
	return nil
}

func (store Neo4jStore) GetAppointment(id string) (*Appointment, error) {
	sess := store.neo4j.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer sess.Close()

	result, err := sess.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		result, err := tx.Run(
			`MATCH (a:Appointment { id:$id })-[r]-(n)
			RETURN *
			`,
			map[string]interface{}{
				"id": id,
			},
		)
		if err != nil {
			return nil, err
		}

		return result.Collect()
	})
	if result == nil || err != nil {
		return nil, err
	}
	records := result.([]*neo4j.Record)

	m := map[string]*Appointment{}
	for _, record := range records {
		processAppointmentRecord(record, m)
	}

	app := m[id]
	return app, nil
}

func processAppointmentRecord(record *neo4j.Record, appointments map[string]*Appointment) {
	appointmentNode := record.Values[0].(neo4j.Node)
	appointmentID := appointmentNode.Props["id"].(string)
	appointment, ok := appointments[appointmentID] // get or create
	if !ok {
		appointment = &Appointment{
			ResourceTypeAndID: ResourceTypeAndID{
				ResourceID:   appointmentID,
				ResourceType: "Appointment",
			},
			Status:      appointmentNode.Props["status"].(string),
			Description: appointmentNode.Props["type"].(string),
		}
		appointments[appointmentID] = appointment
	}

	relationship, _ := record.Get("r")
	node, _ := record.Get("n")
	nodeID := node.(neo4j.Node).Props["id"].(string)

	switch relationship.(neo4j.Relationship).Type {
	case "SUBJECT":
		appointment.Subject = Reference{
			ResourceID:   nodeID,
			ResourceType: "Patient",
		}
	case "ACTOR":
		appointment.Actor = Reference{
			ResourceID:   nodeID,
			ResourceType: "Doctor",
		}
	case "FEEDBACK":
		appointment.Feedback = &Reference{
			ResourceID:   nodeID,
			ResourceType: "Feedback",
		}
	case "APPOINTMENT":
		appointment.Diagnosis = Diagnosis{
			ResourceTypeAndID: ResourceTypeAndID{
				ResourceID:   nodeID,
				ResourceType: "Diagnosis",
			},
			Status: node.(neo4j.Node).Props["status"].(string),
			Name:   node.(neo4j.Node).Props["name"].(string),
			Appointment: Reference{
				ResourceID:   appointmentID,
				ResourceType: "Appointment",
			},
		}
	}
}

func (store Neo4jStore) WriteDiagnosis(d Diagnosis) error {
	sess := store.neo4j.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer sess.Close()
	_, err := sess.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		return tx.Run(
			`
				MERGE (a:Appointment { id:$appointmentId })
				MERGE (d:Diagnosis {
					id: $id,
					status: $status,
					name: $name
				} )
				MERGE (d)-[:APPOINTMENT]-(a)
				RETURN d`,
			map[string]interface{}{
				"id":            d.ID(),
				"status":        d.Status,
				"name":          d.Name,
				"appointmentId": d.Appointment.ResourceID,
			},
		)
	})
	if err != nil {
		return errors.Wrap(err, "problem saving patient "+d.ID())
	}

	return nil
}
