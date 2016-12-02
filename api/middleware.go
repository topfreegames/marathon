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

package api

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/getsentry/raven-go"
	"github.com/labstack/echo"
	"github.com/topfreegames/marathon/log"
	"github.com/uber-go/zap"
)

//NewVersionMiddleware with API version
func NewVersionMiddleware() *VersionMiddleware {
	return &VersionMiddleware{
		Version: VERSION,
	}
}

//VersionMiddleware inserts the current version in all requests
type VersionMiddleware struct {
	Version string
}

// Serve serves the middleware
func (v *VersionMiddleware) Serve(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderServer, fmt.Sprintf("Marathon/v%s", v.Version))
		c.Response().Header().Set("Marathon-Server", fmt.Sprintf("Marathon/v%s", v.Version))
		return next(c)
	}
}

//NewSentryMiddleware returns a new sentry middleware
func NewSentryMiddleware(app *Application) *SentryMiddleware {
	return &SentryMiddleware{
		Application: app,
	}
}

//SentryMiddleware is responsible for sending all exceptions to sentry
type SentryMiddleware struct {
	Application *Application
}

// Serve serves the middleware
func (s *SentryMiddleware) Serve(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := next(c)
		if err != nil {
			if httpErr, ok := err.(*echo.HTTPError); ok {
				if httpErr.Code < 500 {
					return err
				}
			}
			tags := map[string]string{
				"source": "app",
				"type":   "Internal server error",
				"url":    c.Request().URL.RequestURI(),
				"status": fmt.Sprintf("%d", c.Response().Status),
			}
			raven.SetHttpContext(newHTTPFromCtx(c))
			raven.CaptureError(err, tags)
		}
		return err
	}
}

func getHTTPParams(ctx echo.Context) (string, map[string]string, string) {
	qs := ""
	if len(ctx.QueryParams()) > 0 {
		qsBytes, _ := json.Marshal(ctx.QueryParams())
		qs = string(qsBytes)
	}

	headers := map[string]string{}
	for headerKey, headerList := range ctx.Response().Header() {
		headers[string(headerKey)] = headerList[0]
	}

	cookies := string(ctx.Response().Header().Get("Cookie"))
	return qs, headers, cookies
}

func newHTTPFromCtx(ctx echo.Context) *raven.Http {
	qs, headers, cookies := getHTTPParams(ctx)

	h := &raven.Http{
		Method:  string(ctx.Request().Method),
		Cookies: cookies,
		Query:   qs,
		URL:     ctx.Request().URL.RequestURI(),
		Headers: headers,
	}
	return h
}

//NewRecoveryMiddleware returns a configured middleware
func NewRecoveryMiddleware(onError func(error, []byte)) *RecoveryMiddleware {
	return &RecoveryMiddleware{
		OnError: onError,
	}
}

//RecoveryMiddleware recovers from errors
type RecoveryMiddleware struct {
	OnError func(error, []byte)
}

//Serve executes on error handler when errors happen
func (r *RecoveryMiddleware) Serve(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		defer func() {
			if err := recover(); err != nil {
				eError, ok := err.(error)
				if !ok {
					eError = fmt.Errorf(fmt.Sprintf("%v", err))
				}
				if r.OnError != nil {
					r.OnError(eError, debug.Stack())
				}
				c.Error(eError)
			}
		}()
		return next(c)
	}
}

// NewLoggerMiddleware returns the logger middleware
func NewLoggerMiddleware(theLogger zap.Logger) *LoggerMiddleware {
	l := &LoggerMiddleware{Logger: theLogger}
	return l
}

//LoggerMiddleware is responsible for logging to Zap all requests
type LoggerMiddleware struct {
	Logger zap.Logger
}

// Serve serves the middleware
func (l *LoggerMiddleware) Serve(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		l := l.Logger.With(
			zap.String("source", "request"),
		)

		//all except latency to string
		var ip, method, path string
		var status int
		var latency time.Duration
		var startTime, endTime time.Time

		path = c.Path()
		method = c.Request().Method

		startTime = time.Now()

		err := next(c)

		//no time.Since in order to format it well after
		endTime = time.Now()
		latency = endTime.Sub(startTime)

		status = c.Response().Status
		ip = c.Request().RemoteAddr

		route := c.Path()
		reqLog := l.With(
			zap.String("route", route),
			zap.Time("endTime", endTime),
			zap.Int("statusCode", status),
			zap.Duration("latency", latency),
			zap.String("ip", ip),
			zap.String("method", method),
			zap.String("path", path),
		)

		//request failed
		if status > 399 && status < 500 {
			log.W(reqLog, "Request failed.")
			return err
		}

		//request is ok, but server failed
		if status > 499 {
			log.E(reqLog, "Response failed.")
			return err
		}

		//Everything went ok
		if cm := reqLog.Check(zap.InfoLevel, "Request successful."); cm.OK() {
			cm.Write()
		}
		return err
	}
}

//NewNewRelicMiddleware returns the logger middleware
func NewNewRelicMiddleware(app *Application, theLogger zap.Logger) *NewRelicMiddleware {
	l := &NewRelicMiddleware{App: app, Logger: theLogger}
	return l
}

// NewRelicMiddleware is responsible for logging to Zap all requests
type NewRelicMiddleware struct {
	App    *Application
	Logger zap.Logger
}

// Serve serves the middleware
func (nr *NewRelicMiddleware) Serve(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		method := c.Request().Method
		route := fmt.Sprintf("%s %s", method, c.Path())
		txn := nr.App.NewRelic.StartTransaction(route, nil, nil)
		c.Set("txn", txn)
		defer func() {
			c.Set("txn", nil)
			txn.End()
		}()

		err := next(c)
		if err != nil {
			txn.NoticeError(err)
			return err
		}

		return nil
	}
}

// AuthMiddleware automatically adds a version header to response
type AuthMiddleware struct {
	App *Application
}

// Serve Validate that a user exists
func (a AuthMiddleware) Serve(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userEmail := c.Request().Header.Get("x-forwarded-email")
		if userEmail == "" {
			return c.String(401, "Authentication failed.")
		}
		c.Set("user-email", userEmail)
		return next(c)
	}
}

// NewAuthMiddleware returns a configured auth middleware
func NewAuthMiddleware(app *Application) *AuthMiddleware {
	return &AuthMiddleware{
		App: app,
	}
}
