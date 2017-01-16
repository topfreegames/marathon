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
// For more info on APNS payload building, refer to this document:
// https://developer.apple.com/library/content/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/CreatingtheNotificationPayload.html#//apple_ref/doc/uid/TP40008194-CH10-SW1
type APNSMessage struct {
	DeviceToken string                 `json:"DeviceToken"`
	Payload     APNSPayloadContent     `json:"Payload"`
	PushExpiry  int64                  `json:"push_expiry"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// APNSPayloadContent stores payload content of apns message
type APNSPayloadContent struct {
	Aps map[string]interface{} `json:"aps"`
	M   map[string]interface{} `json:"m,omitempty"`
}

// NewAPNSMessage builds an APNSMessage
func NewAPNSMessage(deviceToken string, pushExpiry int64, aps, messageMetadata map[string]interface{}, pushMetadata map[string]interface{}) *APNSMessage {
	if pushMetadata == nil {
		pushMetadata = map[string]interface{}{}
	}

	msg := &APNSMessage{
		DeviceToken: deviceToken,
		PushExpiry:  pushExpiry,
		Metadata:    pushMetadata,
	}
	if aps == nil {
		aps = map[string]interface{}{}
	}
	if messageMetadata == nil {
		messageMetadata = map[string]interface{}{}
	}

	msg.Payload = APNSPayloadContent{
		Aps: aps,
		M:   messageMetadata,
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
