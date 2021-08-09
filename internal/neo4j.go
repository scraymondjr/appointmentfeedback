package internal

import (
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/pkg/errors"
)

func NewNeo4jStore() Neo4jStore {
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

func (store Neo4jStore) WriteResource(r Resource) error {
	switch r := r.(type) {
	case Bundle:
		// TODO guard against number of resources allowed to be written in one bundle
		// TODO include all bundled resources in one txn?
		for _, bundledResource := range r.Resources {
			if err := store.WriteResource(bundledResource); err != nil {
				// TODO decide if should error-fast or continue to loop through all remaining resources
				return errors.Wrap(err, "problem writing resource "+bundledResource.ID())
			}
		}
	case Patient:
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
					"id":         r.ID(),
					"givenName":  r.Name[0].Given[0],
					"familyName": r.Name[0].Family,
				},
			)
		})
		if err != nil {
			return errors.Wrap(err, "problem saving patient "+r.ID())
		}
	case Doctor:
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
					"id":         r.ID(),
					"givenName":  r.Name[0].Given[0],
					"familyName": r.Name[0].Family,
				})
		})
		if err != nil {
			return errors.Wrap(err, "problem saving patient "+r.ID())
		}
	case Appointment:
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
					"id":        r.ID(),
					"status":    r.Status,
					"type":      r.Description,
					"patientId": r.Subject.ResourceID,
					"doctorId":  r.Actor.ResourceID,
				},
			)
		})
		if err != nil {
			return errors.Wrap(err, "problem saving patient "+r.ID())
		}
	case Diagnosis:
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
					"id":            r.ID(),
					"status":        r.Status,
					"name":          r.Name,
					"appointmentId": r.Appointment.ResourceID,
				},
			)
		})
		if err != nil {
			return errors.Wrap(err, "problem saving patient "+r.ID())
		}
	default:
		return errors.Errorf("unknown resource type " + r.Type())
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
