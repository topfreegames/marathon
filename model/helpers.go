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
	"fmt"
	"strings"

	pg "github.com/go-pg/pg/v10"
	"github.com/topfreegames/marathon/interfaces"
)

// InvalidField returns an error telling that field is invalid
func InvalidField(field string) error {
	return fmt.Errorf("invalid %s", field)
}

// GetJobInfoAndApp get the app and the job from the database
// job.ID must be set
func (j *Job) GetJobInfoAndApp(db interfaces.DB) error {
	return db.Model(j).Column("job.*", "App").Where("job.id = ?", j.ID).Select()
}

// GetJobTemplatesByNameAndLocale ...
func (j *Job) GetJobTemplatesByNameAndLocale(db interfaces.DB) (map[string]map[string]Template, error) {
	var templates []Template
	var err error
	if len(strings.Split(j.TemplateName, ",")) > 1 {
		err = db.Model(&templates).Where(
			"app_id = ? AND name IN (?)",
			j.App.ID,
			pg.In(strings.Split(j.TemplateName, ",")),
		).Select()
	} else {
		err = db.Model(&templates).Where(
			"app_id = ? AND name = ?",
			j.App.ID,
			j.TemplateName,
		).Select()
	}
	if err != nil {
		return nil, err
	}
	templateByLocale := make(map[string]map[string]Template)
	for _, tpl := range templates {
		if templateByLocale[tpl.Name] != nil {
			templateByLocale[tpl.Name][tpl.Locale] = tpl
		} else {
			templateByLocale[tpl.Name] = map[string]Template{
				tpl.Locale: tpl,
			}
		}
	}

	if len(templateByLocale) == 0 {
		return nil, fmt.Errorf("No templates were found with name %s and %s", j.TemplateName, j.App.ID)
	}
	return templateByLocale, nil
}
