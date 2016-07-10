package models

import (
	"time"

	"github.com/satori/go.uuid"

	"gopkg.in/gorp.v1"
)

// Notifier identifies uniquely one notifier
type Notifier struct {
	ID        uuid.UUID `db:"id"`
	AppID     uuid.UUID `db:"app_id"`
	Service   string    `db:"service"`
	CreatedAt int64     `db:"created_at"`
	UpdatedAt int64     `db:"updated_at"`
}

// PreInsert populates fields before inserting a new noifier
func (n *Notifier) PreInsert(s gorp.SqlExecutor) error {
	n.ID = uuid.NewV4()
	n.CreatedAt = time.Now().Unix()
	n.UpdatedAt = n.CreatedAt
	return nil
}

// PreUpdate populates fields before updating an noifier
func (n *Notifier) PreUpdate(s gorp.SqlExecutor) error {
	n.UpdatedAt = time.Now().Unix()
	return nil
}

// GetNotifierByID returns an noifier by id
func GetNotifierByID(db DB, id uuid.UUID) (*Notifier, error) {
	obj, err := db.Get(Notifier{}, id)
	if err != nil || obj == nil {
		return nil, &ModelNotFoundError{"Notifier", "id", id}
	}
	return obj.(*Notifier), nil
}

// GetNotifiersByService returns all notifiers for a service
func GetNotifiersByService(db DB, service string) ([]Notifier, error) {
	var notifiers []Notifier
	_, err := db.Select(&notifiers, "SELECT * FROM notifiers WHERE service=$1", service)
	if err != nil || &notifiers == nil || len(notifiers) == 0 {
		return nil, &ModelNotFoundError{"Notifier", "service", service}
	}
	return notifiers, nil
}

// GetNotifiersByApp returns all notifiers for an appID
func GetNotifiersByApp(db DB, appID uuid.UUID) ([]Notifier, error) {
	var notifiers []Notifier
	_, err := db.Select(&notifiers, "SELECT * FROM notifiers WHERE app_id=$1", appID)
	if err != nil || &notifiers == nil || len(notifiers) == 0 {
		return nil, &ModelNotFoundError{"Notifier", "appID", appID}
	}
	return notifiers, nil
}

// GetNotifierByAppAndService returns all notifiers for a service and appID
func GetNotifierByAppAndService(db DB, appID uuid.UUID, service string) (*Notifier, error) {
	var notifiers []Notifier
	_, err := db.Select(&notifiers, "SELECT * FROM notifiers WHERE app_id=$1 AND service=$2", appID, service)
	if err != nil || &notifiers == nil || len(notifiers) == 0 {
		return nil, &ModelNotFoundError{"Notifier", "service", service}
	}
	return &notifiers[0], nil
}

// CreateNotifier creates a new Notifier
func CreateNotifier(db DB, AppID uuid.UUID, Service string) (*Notifier, error) {
	notifier := &Notifier{
		AppID:   AppID,
		Service: Service,
	}
	err := db.Insert(notifier)
	if err != nil {
		return nil, err
	}
	return notifier, nil
}

// UpdateNotifier updates an Notifier
func UpdateNotifier(db DB, id uuid.UUID, AppID uuid.UUID, Service string) (*Notifier, error) {
	notifier, getNotifierErr := GetNotifierByID(db, id)
	if getNotifierErr != nil {
		return nil, getNotifierErr
	}

	notifier.AppID = AppID
	notifier.Service = Service

	_, updateErr := db.Update(notifier)
	if updateErr != nil {
		return nil, updateErr
	}

	return notifier, nil
}
