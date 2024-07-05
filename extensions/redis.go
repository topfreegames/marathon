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

	"github.com/spf13/viper"
	"github.com/topfreegames/marathon/log"
	"github.com/uber-go/zap"
	"gopkg.in/redis.v5"
)

// RedisConnectionError with SourceError that originated the connection failure
type RedisConnectionError struct {
	SourceError error
}

func (e *RedisConnectionError) Error() string {
	return fmt.Sprintf("Could not connect to redis using supplied configuration: %s", e.SourceError.Error())
}

// NewRedis connection with the specified configuration
func NewRedis(prefix string, conf *viper.Viper, logger zap.Logger) (*redis.Client, error) {
	redisHost := conf.GetString(fmt.Sprintf("%s.redis.host", prefix))
	redisPort := conf.GetInt(fmt.Sprintf("%s.redis.port", prefix))
	redisPass := conf.GetString(fmt.Sprintf("%s.redis.pass", prefix))
	redisDB := conf.GetInt(fmt.Sprintf("%s.redis.db", prefix))
	tlsEnabled := conf.GetBool(fmt.Sprintf("%s.redis.tlsEnabled", prefix))

	l := logger.With(
		zap.String("source", "redisExtension"),
		zap.String("operation", "NewRedis"),
		zap.String("redisHost", redisHost),
		zap.Int("redisPort", redisPort),
		zap.Int("redisDB", redisDB),
	)
	log.D(l, "Connecting to redis...")
	opt := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisHost, redisPort),
		Password: redisPass,
		DB:       redisDB,
	}
	if tlsEnabled {
		opt.TLSConfig = &tls.Config{}
	}
	client := redis.NewClient(opt)
	_, err := client.Ping().Result()
	if err != nil {
		log.E(l, "Connection to redis failed.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return nil, &RedisConnectionError{SourceError: err}
	}
	log.I(l, "Connected to redis successfully.")
	return client, nil
}
