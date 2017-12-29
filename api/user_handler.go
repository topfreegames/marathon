/*
 * Copyright (c) 2017 TFG Co <backend@tfgco.com>
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

package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo"
	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/marathon/log"
	"github.com/topfreegames/marathon/model"
	"github.com/uber-go/zap"
)

// ListUsersHandler is the method called when a get to /users is called
func (a *Application) ListUsersHandler(c echo.Context) error {
	l := a.Logger.With(
		zap.String("source", "userHandler"),
		zap.String("operation", "listUsers"),
	)
	users := []model.User{}
	err := WithSegment("db-select", c, func() error {
		return a.DB.Model(&users).Select()
	})
	if err != nil {
		log.E(l, "Failed to list users.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error()})
	}
	log.D(l, "Listed users successfully.", func(cm log.CM) {
		cm.Write(zap.Object("users", users))
	})
	return c.JSON(http.StatusOK, users)
}

//CreateUserHandler is the method called when a post to /users is called
func (a *Application) CreateUserHandler(c echo.Context) error {
	l := a.Logger.With(
		zap.String("source", "userHandler"),
		zap.String("operation", "createUser"),
	)
	createdByEmail := c.Get("user-email").(string)

	user := &model.User{
		ID:        uuid.NewV4(),
		CreatedAt: time.Now().UnixNano(),
		UpdatedAt: time.Now().UnixNano(),
		CreatedBy: createdByEmail,
	}
	err := WithSegment("decodeAndValidate", c, func() error {
		return decodeAndValidate(c, user)
	})
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error(), Value: user})
	}
	err = WithSegment("db-insert", c, func() error {
		return a.DB.Insert(&user)
	})

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return c.JSON(http.StatusConflict, &Error{Reason: err.Error(), Value: user})
		}
		log.E(l, "Failed to create user.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: user})
	}
	log.D(l, "Created user successfully.", func(cm log.CM) {
		cm.Write(zap.Object("user", user))
	})
	return c.JSON(http.StatusCreated, user)
}

//GetUserHandler is the method called when a get to /users/:id is called
func (a *Application) GetUserHandler(c echo.Context) error {
	l := a.Logger.With(
		zap.String("source", "userHandler"),
		zap.String("operation", "getUser"),
		zap.String("userId", c.Param("uid")),
	)
	id, err := uuid.FromString(c.Param("uid"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error()})
	}
	user := &model.User{ID: id}
	err = WithSegment("db-select", c, func() error {
		return a.DB.Select(&user)
	})
	if err != nil {
		if err.Error() == RecordNotFoundString {
			return c.JSON(http.StatusNotFound, map[string]string{})
		}
		log.E(l, "Failed to retrieve user.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: user})
	}
	log.D(l, "Retrieved user successfully.", func(cm log.CM) {
		cm.Write(zap.Object("user", user))
	})
	return c.JSON(http.StatusOK, user)
}

//UpdateUserHandler is the method called when a put to /users/:id is called
func (a *Application) UpdateUserHandler(c echo.Context) error {
	l := a.Logger.With(
		zap.String("source", "userHandler"),
		zap.String("operation", "putUser"),
		zap.String("userId", c.Param("uid")),
	)
	user := &model.User{}
	email := c.Get("user-email").(string)
	user.CreatedBy = email
	user.UpdatedAt = time.Now().UnixNano()
	err := WithSegment("decodeAndValidate", c, func() error {
		return decodeAndValidate(c, user)
	})
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error(), Value: user})
	}
	id, err := uuid.FromString(c.Param("uid"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error()})
	}
	user.ID = id
	err = WithSegment("db-update", c, func() error {
		_, err = a.DB.Model(&user).Column("is_admin").Column("allowed_apps").Column("updated_at").Returning("*").Update()
		return err
	})
	// FIXME: Ugly fix to remove duplicate elements returned by update
	user.AllowedApps = user.AllowedApps[0 : len(user.AllowedApps)/2]
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return c.JSON(http.StatusConflict, &Error{Reason: err.Error(), Value: user})
		}
		log.E(l, "Failed to update user.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: user})
	}
	log.D(l, "Updated user successfully.", func(cm log.CM) {
		cm.Write(zap.Object("user", user))
	})
	return c.JSON(http.StatusOK, user)
}

//DeleteUserHandler is the method called when a delete to /users/:id is called
func (a *Application) DeleteUserHandler(c echo.Context) error {
	l := a.Logger.With(
		zap.String("source", "userHandler"),
		zap.String("operation", "deleteUser"),
		zap.String("userId", c.Param("uid")),
	)
	id, err := uuid.FromString(c.Param("uid"))
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, &Error{Reason: err.Error()})
	}
	user := &model.User{ID: id}
	err = WithSegment("db-delete", c, func() error {
		return a.DB.Delete(&user)
	})
	if err != nil {
		if err.Error() == RecordNotFoundString {
			return c.JSON(http.StatusNotFound, map[string]string{})
		}
		log.E(l, "Failed to delete user.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return c.JSON(http.StatusInternalServerError, &Error{Reason: err.Error(), Value: user})
	}
	log.D(l, "Deleted user successfully.", func(cm log.CM) {
		cm.Write(zap.Object("user", user))
	})
	return c.JSON(http.StatusNoContent, "")
}
