/*
 * Copyright (c) 2019 TFG Co <backend@tfgco.com>
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

package lib

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	ehttp "github.com/topfreegames/go-extensions-http"
)

var client *http.Client

func getHTTPClient(config *Config) *http.Client {
	if client == nil {
		client = ehttp.New()
		client.Timeout = config.Timeout
	}

	return client
}

func (m *Marathon) sendTo(
	ctx context.Context,
	method, url string,
	payload, response interface{},
) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	var req *http.Request

	if payload != nil {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(payloadJSON))
		if err != nil {
			return err
		}
	} else {
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			return err
		}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-forwarded-email", m.userEmail)
	if ctx == nil {
		ctx = context.Background()
	}
	req = req.WithContext(ctx)

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, respErr := ioutil.ReadAll(resp.Body)
	if respErr != nil {
		return respErr
	}

	if resp.StatusCode > 399 {
		return newRequestError(resp.StatusCode, string(body))
	}

	err = json.Unmarshal(body, response)
	if err != nil {
		return err
	}

	return nil
}

func (m *Marathon) buildURL(pathname string) string {
	return fmt.Sprintf("%s/apps/%s/%s", m.url, m.appID, pathname)
}

func (m *Marathon) buildCreateJobURL(template string) string {
	return m.buildURL(fmt.Sprintf("jobs?template=%s", template))
}

func (m *Marathon) buildListJobsURL(template string) string {
	return m.buildURL(fmt.Sprintf("jobs?template=%s", template))
}
