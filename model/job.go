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
	"github.com/satori/go.uuid"
	"time"
)

// Job is the job model struct
type Job struct {
	ID               uuid.UUID `sql:"type:uuid;default:uuid_generate_v4()" json:"id"`
	TotalBatches     int       `json:"totalBatches"`
	CompletedBatches int       `gorm:"not null" sql:"default:0" json:"completedBatches"`
	CompletedAt      time.Time `json:"completedAt"`
	ExpiresAt        time.Time `json:"expiresAt"`
	Service          string    `json:"service"`
	Filters          string    `sql:"type:JSONB NOT NULL DEFAULT '{}'::JSONB" json:"filters"`
	CsvURL           string    `json:"csvUrl"`
	CreatedBy        string    `json:"createdBy"`
	App              App       `json:"app"`
	AppID            uuid.UUID `sql:"type:uuid" gorm:"not null" json:"appId"`
	Template         Template  `json:"template"`
	TemplateID       uuid.UUID `sql:"type:uuid" gorm:"not null" json:"templateId"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}
