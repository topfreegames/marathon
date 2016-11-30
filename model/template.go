package model

import (
	"github.com/satori/go.uuid"
	"time"
)

// Template is the template model struct
type Template struct {
	ID           uuid.UUID `sql:"type:uuid;default:uuid_generate_v4()"`
	Name         string    `gorm:"not null;unique_index:name_locale_app"`
	Locale       string    `gorm:"not null;unique_index:name_locale_app"`
	Defaults     string    `sql:"type:JSONB NOT NULL DEFAULT '{}'::JSONB"`
	Body         string    `sql:"type:JSONB NOT NULL DEFAULT '{}'::JSONB"`
	CompiledBody string    `gorm:"not null"`
	CreatedBy    string    `gorm:"not null"`
	App          App
	AppID        uuid.UUID `sql:"type:uuid" gorm:"not null;unique_index:name_locale_app"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
