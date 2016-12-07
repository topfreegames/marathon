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

package messages

import "encoding/json"

// APNSMessage might need to update the json encoding if we change to snake case
type APNSMessage struct {
	DeviceToken string             `json:"DeviceToken"`
	Payload     APNSPayloadContent `json:"Payload"`
	PushExpiry  int64              `json:"push_expiry"`
}

// APNSPayloadContent stores payload content of apns message
type APNSPayloadContent struct {
	Aps map[string]interface{} `json:"aps"`
	M   map[string]interface{} `json:"m,omitempty"`
}

// NewAPNSMessage builds an APNSMessage
func NewAPNSMessage(deviceToken string, pushExpiry int64, aps, m map[string]interface{}) *APNSMessage {
	msg := &APNSMessage{
		DeviceToken: deviceToken,
		PushExpiry:  pushExpiry,
	}
	if aps == nil {
		aps = map[string]interface{}{}
	}
	if m == nil {
		m = map[string]interface{}{}
	}
	msg.Payload = APNSPayloadContent{
		Aps: aps,
		M:   m,
	}
	return msg
}

//ToJSON returns the serialized message
func (m *APNSMessage) ToJSON() (string, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
