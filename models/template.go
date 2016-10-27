package models

import (
	"github.com/satori/go.uuid"
	"gopkg.in/gorp.v1"
)

// Template identifies uniquely one template
type Template struct {
	ID       uuid.UUID              `db:"id"`
	Name     string                 `db:"name"`
	Locale   string                 `db:"locale"`
	Defaults map[string]interface{} `db:"defaults"`
	Body     map[string]interface{} `db:"body"`
}

// PreInsert populates fields before inserting a new template
func (o *Template) PreInsert(s gorp.SqlExecutor) error {
	o.ID = uuid.NewV4()
	return nil
}

// GetTemplateByID returns a template by id
func GetTemplateByID(db *DB, id uuid.UUID) (*Template, error) {
	obj, err := db.Get(Template{}, id)
	if err != nil || obj == nil {
		return nil, &ModelNotFoundError{"Template", "id", id}
	}
	return obj.(*Template), nil
}

// GetTemplatesByName returns templates with the given name
func GetTemplatesByName(db *DB, name string) ([]Template, error) {
	var templates []Template
	_, err := db.Select(&templates, "SELECT * FROM templates WHERE name=$1", name)

	if err != nil || &templates == nil || len(templates) == 0 {
		return nil, &ModelNotFoundError{"Template", "name", name}
	}
	return templates, nil
}

// GetTemplateByNameAndLocale returns a template by its name and locale
func GetTemplateByNameAndLocale(db *DB, name string, locale string) (*Template, error) {
	var template Template
	err := db.SelectOne(&template, "SELECT * FROM templates WHERE name=$1 AND locale=$2", name, locale)
	if err != nil && locale != "en" {
		// If no templates with the specified locale were found try en
		err = db.SelectOne(&template, "SELECT * FROM templates WHERE name=$1 AND locale=$2", name, "en")
	}
	if err != nil || &template == nil {
		return nil, &ModelNotFoundError{"Template", "name", name}
	}
	return &template, nil
}

// CreateTemplate creates a new Template
func CreateTemplate(db *DB, name string, locale string, defaults map[string]interface{}, body map[string]interface{}) (*Template, error) {
	template := &Template{
		Name:     name,
		Locale:   locale,
		Defaults: defaults,
		Body:     body,
	}
	err := db.Insert(template)
	if err != nil {
		return nil, err
	}
	return template, nil
}

// UpdateTemplate updates a Template
func UpdateTemplate(db *DB, id uuid.UUID, name string, locale string, defaults map[string]interface{}, body map[string]interface{}) (*Template, error) {
	template, getTemplateErr := GetTemplateByID(db, id)
	if getTemplateErr != nil {
		return nil, getTemplateErr
	}

	template.Name = name
	template.Locale = locale
	template.Defaults = defaults
	template.Body = body

	_, updateErr := db.Update(template)
	if updateErr != nil {
		return nil, updateErr
	}

	return template, nil
}
