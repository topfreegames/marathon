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

import (
	"encoding/json"
)

// GCMMessage is the struct to store a gcm message
// For more info on the GCM Message Data attribute refer to:
// https://developers.google.com/cloud-messaging/concept-options
type GCMMessage struct {
	To                     string                 `json:"to"`
	Data                   map[string]interface{} `json:"data"`
	TimeToLive             int64                  `json:"time_to_live,omitempty"`
	DelayWhileIdle         bool                   `json:"delay_while_idle,omitempty"`
	DeliveryReceiptRequest bool                   `json:"delivery_receipt_requested,omitempty"`
	DryRun                 bool                   `json:"dry_run"`
	MessageID              string                 `json:"message_id"`
	Metadata               map[string]interface{} `json:"metadata"`
}

// NewGCMMessage builds a new GCM Message
func NewGCMMessage(to string, data, messageMetadata map[string]interface{}, pushMetadata map[string]interface{}, timeToLive int64, templateName string) *GCMMessage {
	if data == nil {
		data = map[string]interface{}{}
	}

	data["templateName"] = templateName
	if messageMetadata != nil && len(messageMetadata) > 0 {
		data["m"] = messageMetadata
	}

	msg := &GCMMessage{
		To:                     to,
		Data:                   data,
		TimeToLive:             timeToLive,
		DryRun:                 false,
		DeliveryReceiptRequest: false,
		DelayWhileIdle:         false,
		MessageID:              "",
		Metadata:               pushMetadata,
	}

	return msg
}

//ToJSON returns the serialized message
func (m *GCMMessage) ToJSON() (string, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
