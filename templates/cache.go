package templates

import (
	"fmt"
	"time"

	"git.topfreegames.com/topfreegames/marathon/models"

	"github.com/uber-go/zap"
)

// Cache stores cached templates and the duration the cached templates
// should remain valid
type Cache struct {
	Cells       map[string]*templateCell // Cells uses as key: template_name,locale
	CellTimeout time.Duration
}

type templateCell struct {
	TemplateName string
	Service      string
	Locale       string
	Cell         *models.Template
	Expiry       time.Time
}

// CreateTemplateCache returns a new instance of TemplateCache, timeout is
// given in seconds
func CreateTemplateCache(timeout int64) *Cache {
	return &Cache{
		Cells:       map[string]*templateCell{},
		CellTimeout: time.Duration(timeout) * time.Second,
	}
}

// FindTemplate is a method that returns a pointer to a template if a valid
// cached version exists
func (tc *Cache) FindTemplate(l zap.Logger, name string, service string, locale string) *models.Template {
	key := fmt.Sprintf("%s,%s", name, locale)
	temp := tc.Cells[key]
	if temp != nil {
		// Check if template is still valid
		if temp.Expiry.After(time.Now()) {
			return temp.Cell
		}
		l.Debug("Template expired", zap.String("name", name), zap.String("service", service), zap.String("locale", locale))
		// Expired, delete entry from cache
		delete(tc.Cells, key)
	}
	l.Debug("No valid templates found in cache", zap.String("name", name), zap.String("service", service), zap.String("locale", locale))
	return nil
}

// AddTemplate adds a new template to the cache
func (tc *Cache) AddTemplate(l zap.Logger, name string, service string, locale string, temp *models.Template) {
	key := fmt.Sprintf("%s,%s", name, locale)
	expiry := time.Now().Add(tc.CellTimeout)
	tempCell := &templateCell{
		TemplateName: name,
		Locale:       locale,
		Cell:         temp,
		Expiry:       expiry,
	}
	l.Debug("Added template to template cach", zap.String("name", name), zap.String("service", service), zap.String("locale", locale))
	tc.Cells[key] = tempCell
}
