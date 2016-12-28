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

package extensions

import (
	"fmt"
	"strings"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/spf13/viper"
	"github.com/topfreegames/marathon/log"
	"github.com/uber-go/zap"
)

// SendgridClient is the struct that keeps sendgrid info
type SendgridClient struct {
	Config *viper.Viper
	Logger zap.Logger
	APIKey string
}

// NewSendgridClient creates a new client
func NewSendgridClient(config *viper.Viper, logger zap.Logger, apiKey string) *SendgridClient {
	client := &SendgridClient{
		Config: config,
		Logger: logger,
		APIKey: apiKey,
	}
	client.LoadDefaults()
	return client
}

// LoadDefaults sets default values for keys needed by this module
func (c *SendgridClient) LoadDefaults() {
	c.Config.SetDefault("sendgrid.sender", "no-reply@tfgco.com")
	c.Config.SetDefault("sendgrid.addressees", "")
}

// SendgridSendEmail sends an email with the given message and addressee
func (c *SendgridClient) SendgridSendEmail(addressee, subject, message string) error {
	l := c.Logger.With(
		zap.String("source", "sendgridExtension"),
		zap.String("operation", "SendgridSendEmail"),
	)

	from := mail.NewEmail("Marathon", c.Config.GetString("sendgrid.sender"))
	content := mail.NewContent("text/plain", message)

	var tos []*mail.Email
	defaultAdressees := c.Config.GetString("sendgrid.addressees")
	if defaultAdressees != "" {
		defaultAdresseesAr := strings.Split(defaultAdressees, ",")
		tos = make([]*mail.Email, len(defaultAdresseesAr)+1)
		for idx, email := range defaultAdresseesAr {
			tos[idx] = mail.NewEmail(strings.Split(email, "@")[0], email)
		}
	} else {
		tos = make([]*mail.Email, 1)
	}

	tos[len(tos)-1] = mail.NewEmail(strings.Split(addressee, "@")[0], addressee)

	m := new(mail.SGMailV3)
	m.SetFrom(from)
	m.Subject = subject
	p := mail.NewPersonalization()
	p.AddTos(tos...)
	m.AddPersonalizations(p)
	m.AddContent(content)

	request := sendgrid.GetRequest(c.APIKey, "/v3/mail/send", "https://api.sendgrid.com")
	request.Method = "POST"
	request.Body = mail.GetRequestBody(m)
	log.D(l, "Sending email...")
	response, err := sendgrid.API(request)
	if err != nil {
		log.E(l, "Failed to send email.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
	} else if response.StatusCode >= 400 {
		log.E(l, "Failed to send email.", func(cm log.CM) {
			cm.Write(zap.Error(fmt.Errorf(fmt.Sprintf("Failed with status %s and body %s.", response.StatusCode, response.Body))))
		})
	} else {
		log.I(l, "Sent email successfully.")
	}
	return err
}
