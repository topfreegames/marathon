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
	"crypto/tls"
	"fmt"
	"time"

	pg "github.com/go-pg/pg/v10"

	"github.com/spf13/viper"
	"github.com/topfreegames/marathon/interfaces"
	"github.com/uber-go/zap"
)

// PGClient is the struct that connects to PostgreSQL
type PGClient struct {
	Config *viper.Viper
	DB     interfaces.DB
	Logger zap.Logger
}

// NewPGClient creates a new client
func NewPGClient(prefix string, config *viper.Viper, logger zap.Logger, PGOrNil ...interfaces.DB) (*PGClient, error) {
	client := &PGClient{
		Config: config,
		Logger: logger,
	}
	if len(PGOrNil) > 0 {
		client.Connect(prefix, PGOrNil[0])
	} else {
		err := client.Connect(prefix)
		if err != nil {
			return nil, err
		}
	}

	return client, nil
}

// Connect to PG
func (c *PGClient) Connect(prefix string, PGOrNil ...interfaces.DB) error {
	user := c.Config.GetString(fmt.Sprintf("%s.user", prefix))
	pass := c.Config.GetString(fmt.Sprintf("%s.pass", prefix))
	host := c.Config.GetString(fmt.Sprintf("%s.host", prefix))
	db := c.Config.GetString(fmt.Sprintf("%s.database", prefix))
	port := c.Config.GetInt(fmt.Sprintf("%s.port", prefix))
	poolSize := c.Config.GetInt(fmt.Sprintf("%s.poolSize", prefix))
	maxRetries := c.Config.GetInt(fmt.Sprintf("%s.maxRetries", prefix))
	sslmode := c.Config.GetBool(fmt.Sprintf("%s.sslmode", prefix))

	if len(PGOrNil) > 0 {
		c.DB = PGOrNil[0]
		return nil
	}

	var tlsConfig *tls.Config
	if sslmode {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}
	opts := &pg.Options{
		Addr:       fmt.Sprintf("%s:%d", host, port),
		User:       user,
		Password:   pass,
		Database:   db,
		PoolSize:   poolSize,
		MaxRetries: maxRetries,
		TLSConfig:  tlsConfig,
	}

	conn := pg.Connect(opts)
	c.DB = conn

	return nil
}

// IsConnected determines if the client is connected to PG
func (c *PGClient) IsConnected() bool {
	res, err := c.DB.Exec("select 1")
	if err != nil {
		return false
	}
	return res.RowsReturned() == 1
}

// Close the connections to PG
func (c *PGClient) Close() error {
	c.DB.Close()
	return nil
}

// WaitForConnection loops until postgres is connected
func (c *PGClient) WaitForConnection(timeout int) error {
	start := time.Now().Unix()
	for !c.IsConnected() || time.Now().Unix()-start > int64(timeout) {
		time.Sleep(10 * time.Millisecond)
	}

	if time.Now().Unix()-start > int64(timeout) {
		return fmt.Errorf("Timed out waiting for Postgres to connect")
	}
	return nil
}
