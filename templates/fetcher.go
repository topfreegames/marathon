package templates

import (
	"fmt"

	"git.topfreegames.com/topfreegames/marathon/messages"
	"git.topfreegames.com/topfreegames/marathon/models"

	"github.com/imdario/mergo"
	_ "github.com/lib/pq" // Used by gorp
	"github.com/uber-go/zap"
)

// Fetcher starts a new fetcher worker which reads from inChan and writes to
// outChan, using the db
func Fetcher(l zap.Logger, inChan <-chan *messages.InputMessage, outChan chan<- *messages.TemplatedMessage, doneChan <-chan struct{}, db *models.DB) {
	l.Info("Starting fetcher")
	tc := CreateTemplateCache(60)
	for {
		select {
		case <-doneChan:
			return // breaks out of the for
		case input, ok := <-inChan:
			if !ok {
				l.Error("Not consuming InputMessages", zap.String("msg", fmt.Sprintf("%+v", input)))
				return
			}
			output, err := FetchTemplate(l, input, db, tc)
			if err != nil {
				l.Error("Error processing request message", zap.Error(err))
				continue
			}
			l.Debug("Fetched template", zap.Object("TemplatedMessage", output))
			outChan <- output
		}
	}
}

// FetchTemplate fetches the template with name and locale from the cache and then from db
func FetchTemplate(l zap.Logger, input *messages.InputMessage, db *models.DB, tc *Cache) (*messages.TemplatedMessage, error) {
	templated := messages.NewTemplatedMessage()
	templated.App = input.App
	templated.Token = input.Token
	templated.Service = input.Service
	templated.PushExpiry = input.PushExpiry
	templated.Locale = input.Locale
	templated.Metadata = input.Metadata
	templated.Params = input.Params
	templated.Message = input.Message

	if input.Template != "" {
		// First search the cache for valid templates
		template := tc.FindTemplate(l, input.Template, input.Service, input.Locale)
		if template == nil {
			dbTemplate, err := models.GetTemplateByNameServiceAndLocale(db, input.Template, input.Service, input.Locale)
			if err != nil {
				l.Error("Error fetching template", zap.Error(err))
				return nil, err
			}
			tc.AddTemplate(l, input.Template, input.Service, input.Locale, dbTemplate)
			template = dbTemplate
		}
		mergo.Merge(&input.Params, template.Defaults)
		templated.Params = input.Params
		templated.Locale = template.Locale
		templated.Message = template.Body

		l.Debug("Fetched template", zap.String("name", input.Template), zap.String("locale", input.Locale))
	}
	return templated, nil
}
