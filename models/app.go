package models

import (
	"time"

	"github.com/satori/go.uuid"

	"gopkg.in/gorp.v1"
)

// App identifies uniquely one app
type App struct {
	ID             string `db:"id"`
	Name           string `db:"name"`
	OrganizationID string `db:"organization_id"`
	AppGroup       string `db:"app_group"`
	CreatedAt      int64  `db:"created_at"`
	UpdatedAt      int64  `db:"updated_at"`
	DeletedAt      int64  `db:"deleted_at"`
}

// PreInsert populates fields before inserting a new app
func (a *App) PreInsert(s gorp.SqlExecutor) error {
	a.ID = uuid.NewV4().String()
	a.CreatedAt = time.Now().Unix()
	a.UpdatedAt = a.CreatedAt
	return nil
}

// PreUpdate populates fields before updating an app
func (a *App) PreUpdate(s gorp.SqlExecutor) error {
	a.UpdatedAt = time.Now().Unix()
	return nil
}

// GetAppByID returns an app by id
func GetAppByID(db DB, id string) (*App, error) {
	obj, err := db.Get(App{}, id)
	if err != nil || obj == nil {
		return nil, &ModelNotFoundError{"App", "id", id}
	}
	return obj.(*App), nil
}

// GetAppByName returns an app by its name
func GetAppByName(db DB, name string) (*App, error) {
	var app App
	err := db.SelectOne(&app, "SELECT * FROM apps WHERE name=$1", name)
	if err != nil || &app == nil {
		return nil, &ModelNotFoundError{"App", "name", name}
	}
	return &app, nil
}

// GetAppsByGroup returns all apps in a group
func GetAppsByGroup(db DB, group string) ([]App, error) {
	var apps []App
	_, err := db.Select(&apps, "SELECT * FROM apps WHERE app_group=$1", group)
	if err != nil || &apps == nil || len(apps) == 0 {
		return nil, &ModelNotFoundError{"App", "group", group}
	}
	return apps, nil
}

// CreateApp creates a new App
func CreateApp(db DB, Name string, OrganizationID string, AppGroup string) (*App, error) {
	app := &App{
		Name:           Name,
		OrganizationID: OrganizationID,
		AppGroup:       AppGroup,
	}
	err := db.Insert(app)
	if err != nil {
		return nil, err
	}
	return app, nil
}

// UpdateApp updates an App
func UpdateApp(db DB, id string, Name string, OrganizationID string, AppGroup string) (*App, error) {
	app, getAppErr := GetAppByID(db, id)
	if getAppErr != nil {
		return nil, getAppErr
	}

	app.Name = Name
	app.OrganizationID = OrganizationID
	app.AppGroup = AppGroup

	_, updateErr := db.Update(app)
	if updateErr != nil {
		return nil, updateErr
	}

	return app, nil
}
