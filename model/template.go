/*
 * Copyright (c) 2016 TFG Co <backend@tfgco.com>
 * Author: TFG Co <backend@tfgco.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of
 * this software and associated documentation files (the "Software"), to deal in
 * the Software without restriction, including without limitation the rights to
 * use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
 * the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
 * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package model

import (
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo"
	"github.com/satori/go.uuid"
)

// Template is the template model struct
type Template struct {
	ID           uuid.UUID `sql:"type:uuid;default:uuid_generate_v4()" json:"id"`
	Name         string    `gorm:"not null;unique_index:name_locale_app" json:"name"`
	Locale       string    `gorm:"not null;unique_index:name_locale_app" json:"locale"`
	Defaults     string    `sql:"type:JSONB NOT NULL DEFAULT '{}'::JSONB" json:"defaults"`
	Body         string    `sql:"type:JSONB NOT NULL DEFAULT '{}'::JSONB" json:"body"`
	CompiledBody string    `gorm:"not null" json:"compiledBody"`
	CreatedBy    string    `gorm:"not null" json:"createdBy"`
	App          App       `json:"app"`
	AppID        uuid.UUID `sql:"type:uuid" gorm:"not null;unique_index:name_locale_app" json:"appId"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// Validate implementation of the InputValidation interface
func (t *Template) Validate(c echo.Context) error {
	valid := govalidator.StringLength(t.Name, "1", "255")
	if !valid {
		return InvalidField("name")
	}
	valid = govalidator.StringLength(t.Locale, "1", "10")
	if !valid {
		return InvalidField("locale")
	}
	valid = govalidator.IsJSON(t.Defaults)
	if !valid {
		return InvalidField("defaults")
	}
	valid = govalidator.IsJSON(t.Body)
	if !valid {
		return InvalidField("body")
	}
	valid = govalidator.IsEmail(t.CreatedBy)
	if !valid {
		return InvalidField("createdBy")
	}
	return nil
}
