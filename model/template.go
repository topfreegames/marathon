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
	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo/v4"
	"github.com/satori/go.uuid"
)

// Template is the template model struct
type Template struct {
	ID        uuid.UUID              `sql:",pk" json:"id"`
	Name      string                 `json:"name"`
	Locale    string                 `json:"locale"`
	Defaults  map[string]interface{} `json:"defaults"`
	Body      map[string]interface{} `json:"body"`
	CreatedBy string                 `json:"createdBy"`
	App       App                    `json:"app"`
	AppID     uuid.UUID              `json:"appId"`
	CreatedAt int64                  `json:"createdAt"`
	UpdatedAt int64                  `json:"updatedAt"`
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
	valid = len(t.Body) > 0
	if !valid {
		return InvalidField("body")
	}
	return nil
}
