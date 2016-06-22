package api

import (
	"git.topfreegames.com/topfreegames/marathon/models"
	"github.com/kataras/iris"
	"gopkg.in/gorp.v1"
)

// TransactionMiddleware wraps transactions around the request
type TransactionMiddleware struct {
	Application *Application
}

// Serve Automatically wrap transaction around the request
func (m *TransactionMiddleware) Serve(c *iris.Context) {
	c.Set("db", m.Application.Db)

	tx, err := (m.Application.Db).(*gorp.DbMap).Begin()
	if err == nil {
		c.Set("db", tx)
		c.Next()

		if c.Response.StatusCode() > 399 {
			tx.Rollback()
			return
		}

		tx.Commit()
		c.Set("db", m.Application.Db)
	}
}

// GetCtxDB returns the proper database connection depending on the request context
func GetCtxDB(ctx *iris.Context) models.DB {
	return ctx.Get("db").(models.DB)
}
