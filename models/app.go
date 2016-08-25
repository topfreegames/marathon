package models

import (
	"time"

	"github.com/satori/go.uuid"
	"gopkg.in/gorp.v1"
)

// App identifies uniquely one app
type App struct {
	ID             uuid.UUID `db:"id"`
	Name           string    `db:"name"`
	OrganizationID uuid.UUID `db:"organization_id"`
	AppGroup       string    `db:"group"`
	CreatedAt      int64     `db:"created_at"`
	UpdatedAt      int64     `db:"updated_at"`
}

// AppNotifier identifies an app/notifier
type AppNotifier struct {
	AppID             uuid.UUID `json:"appID"`
	AppOrganizationID uuid.UUID `json:"appOrganizationID"`
	AppName           string    `json:"appName"`
	AppGroup          string    `json:"appGroup"`
	AppCreatedAt      int64     `json:"appCreatedAt"`
	AppUpdatedAt      int64     `json:"appUpdatedAt"`
	NotifierID        uuid.UUID `json:"notifierID"`
	NotifierAppID     uuid.UUID `json:"notifierAppID"`
	NotifierService   string    `json:"notifierService"`
	NotifierCreatedAt int64     `json:"notifierCreatedAt"`
	NotifierUpdatedAt int64     `json:"notifierUpdatedAt"`
}

// PreInsert populates fields before inserting a new app
func (a *App) PreInsert(s gorp.SqlExecutor) error {
	a.ID = uuid.NewV4()
	a.CreatedAt = time.Now().Unix()
	a.UpdatedAt = a.CreatedAt
	return nil
}

// PreUpdate populates fields before updating an app
func (a *App) PreUpdate(s gorp.SqlExecutor) error {
	a.UpdatedAt = time.Now().Unix()
	return nil
}

// CountApps count the number of apps in the db
func CountApps(db *DB) (int64, error) {
	count, err := db.SelectInt("SELECT COUNT(1) FROM apps")
	if err != nil {
		return int64(0), err
	}
	return count, nil
}

// GetAppByID returns an app by id
func GetAppByID(db *DB, id uuid.UUID) (*App, error) {
	obj, err := db.Get(App{}, id)
	if err != nil || obj == nil {
		return nil, &ModelNotFoundError{"App", "id", id}
	}
	return obj.(*App), nil
}

// GetAppByName returns an app by its name
func GetAppByName(db *DB, name string) (*App, error) {
	var app App
	err := db.SelectOne(&app, "SELECT * FROM apps WHERE name=$1", name)
	if err != nil || &app == nil {
		return nil, &ModelNotFoundError{"App", "name", name}
	}
	return &app, nil
}

// GetAppsByGroup returns all apps in a group
func GetAppsByGroup(db *DB, group string) ([]App, error) {
	var apps []App
	_, err := db.Select(&apps, "SELECT * FROM apps WHERE \"group\"=$1", group)
	if err != nil || &apps == nil || len(apps) == 0 {
		return nil, &ModelNotFoundError{"App", "group", group}
	}
	return apps, nil
}

// GetAppNotifiers returns all apps with notifiers
func GetAppNotifiers(db *DB) ([]AppNotifier, error) {
	query := `SELECT
    a.id AS AppID,
    a.organization_id AS AppOrganizationID,
    a.name AS AppName,
    a.group AS AppGroup,
    a.created_at AS AppCreatedAt,
    a.updated_at AS AppUpdatedAt,
    n.id AS NotifierID,
    n.app_id AS NotifierAppID,
    n.service AS NotifierService,
    n.created_at AS NotifierCreatedAt,
    n.updated_at AS NotifierUpdatedAt
  FROM
    apps a
  INNER JOIN notifiers n ON n.app_id=a.id
  `
	var appNotifiers []AppNotifier
	_, err := db.Select(&appNotifiers, query)
	if err != nil {
		return nil, err
	}
	if &appNotifiers == nil {
		return nil, &ModelNotFoundError{"App", "", ""}
	}

	return appNotifiers, nil
}

// CreateApp creates a new App
func CreateApp(db *DB, name string, organizationid uuid.UUID, appgroup string) (*App, error) {
	app := &App{
		Name:           name,
		OrganizationID: organizationid,
		AppGroup:       appgroup,
	}
	err := db.Insert(app)
	if err != nil {
		return nil, err
	}
	return app, nil
}

// UpdateApp updates an App
func UpdateApp(db *DB, id uuid.UUID, name string, organizationid uuid.UUID, appgroup string) (*App, error) {
	app, getAppErr := GetAppByID(db, id)
	if getAppErr != nil {
		return nil, getAppErr
	}

	app.Name = name
	app.OrganizationID = organizationid
	app.AppGroup = appgroup

	_, updateErr := db.Update(app)
	if updateErr != nil {
		return nil, updateErr
	}

	return app, nil
}
