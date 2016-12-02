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

// App is the app model struct
type App struct {
	ID        uuid.UUID `sql:"type:uuid;default:uuid_generate_v4()" json:"id"`
	Key       string    `gorm:"not null" json:"key"`
	BundleID  string    `gorm:"unique_index;not null" json:"bundleId"`
	CreatedBy string    `gorm:"not null" json:"createdBy"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Validate implementation of the InputValidation interface
func (a *App) Validate(c echo.Context) error {
	valid := govalidator.StringLength(a.Key, "1", "255")
	if !valid {
		return InvalidField("key")
	}
	valid = govalidator.StringMatches(a.BundleID, "^[a-z0-9]+\\.[a-z0-9]+\\.[a-z0-9]+$")
	if !valid {
		return InvalidField("bundleId")
	}
	valid = govalidator.IsEmail(a.CreatedBy)
	if !valid {
		return InvalidField("createdBy")
	}
	return nil
}
