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
	"fmt"
	"strconv"
	"strings"
)

// Notification mirrors the notification block consumed by pusher's firebase
// message handler (interfaces.Message.Notification). pusher's buildIOSMessage
// builds the APNs `aps.alert` ONLY from these top-level fields — never from
// `data` — so the FCM-iOS dispatch path MUST populate this. This is the
// asymmetry with the Android `gcm` path: Android renders data messages
// client-side, iOS needs a real `aps.alert` or the push is silent.
type Notification struct {
	Title        string `json:"title,omitempty"`
	Body         string `json:"body,omitempty"`
	Sound        string `json:"sound,omitempty"`
	Badge        string `json:"badge,omitempty"`
	BodyLocKey   string `json:"body_loc_key,omitempty"`
	BodyLocArgs  string `json:"body_loc_args,omitempty"`
	TitleLocKey  string `json:"title_loc_key,omitempty"`
	TitleLocArgs string `json:"title_loc_args,omitempty"`
	ImageURL     string `json:"image_url,omitempty"`
}

// FCMMessage is the wire shape pusher's firebase consumer expects on the
// push-<game>_ios-<size> topic: a generic message carrying a TOP-LEVEL
// notification (interfaces.Message + kafkaFCMMessage's metadata/push_expiry).
//
// It is deliberately NOT a GCMMessage. GCMMessage puts the whole template under
// `data`, and pusher's buildIOSMessage ignores `data` when building `aps`, so a
// GCMMessage delivered on the _ios topic produces an empty `aps` (a silent
// push). FCMMessage carries the rendered alert in `notification` so iOS renders
// a banner.
type FCMMessage struct {
	To               string                 `json:"to"`
	Notification     *Notification          `json:"notification,omitempty"`
	ContentAvailable bool                   `json:"content_available,omitempty"`
	Data             map[string]interface{} `json:"data,omitempty"`
	DryRun           bool                   `json:"dry_run"`
	PushExpiry       int64                  `json:"push_expiry,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// NewFCMMessage builds an FCMMessage from an APNs-shaped template payload (the
// same payload the apns path feeds into Payload.aps). The APNs alert/sound/
// badge/content-available are mapped onto pusher's notification block; custom
// message data (`m`) and templateName are carried in `data` to match the gcm
// path. `pushExpiry` is given in seconds (expiresAt/1e9) and converted here.
func NewFCMMessage(to string, payload, messageMetadata, pushMetadata map[string]interface{}, pushExpiry int64, templateName string) *FCMMessage {
	if payload == nil {
		payload = map[string]interface{}{}
	}

	data := map[string]interface{}{
		"templateName": templateName,
	}
	if len(messageMetadata) > 0 {
		data["m"] = messageMetadata
	}

	msg := &FCMMessage{
		To:               to,
		Notification:     apsAlertToNotification(payload),
		ContentAvailable: apsContentAvailable(payload),
		Data:             data,
		DryRun:           false,
		Metadata:         pushMetadata,
	}

	// pusher's firebase handler drops messages whose push_expiry has already
	// passed, comparing against MakeTimestamp() which is in MILLISECONDS.
	// marathon's pushExpiry is absolute unix SECONDS, so convert. A zero value
	// means "no expiry" and is left unset so the check is skipped — exactly how
	// the Android gcm path (which has no push_expiry field) behaves.
	if pushExpiry > 0 {
		msg.PushExpiry = pushExpiry * 1000
	}

	return msg
}

// apsAlertToNotification maps an APNs aps payload onto pusher's notification
// block. APNs `alert` may be a plain string (the body) or a dict with
// title/body/loc keys. `sound` may be a string or a critical-alert dict.
// Options pusher's Notification cannot represent (mutable-content, thread-id,
// category, interruption-level, subtitle) are dropped — closing that gap is a
// pusher-side change to interfaces.Message/buildIOSMessage.
func apsAlertToNotification(payload map[string]interface{}) *Notification {
	n := &Notification{}

	switch alert := payload["alert"].(type) {
	case string:
		n.Body = alert
	case map[string]interface{}:
		n.Title = stringField(alert, "title")
		n.Body = stringField(alert, "body")
		n.TitleLocKey = stringField(alert, "title-loc-key")
		n.BodyLocKey = stringField(alert, "loc-key")
		n.TitleLocArgs = joinLocArgs(alert["title-loc-args"])
		n.BodyLocArgs = joinLocArgs(alert["loc-args"])
	}

	switch sound := payload["sound"].(type) {
	case string:
		n.Sound = sound
	case map[string]interface{}:
		// Critical-alert sound dictionary: {critical, name, volume}.
		n.Sound = stringField(sound, "name")
	}

	if badge, ok := payload["badge"]; ok {
		n.Badge = numberToString(badge)
	}

	if *n == (Notification{}) {
		return nil
	}
	return n
}

// apsContentAvailable reports whether the aps payload requests a
// content-available (silent/background) push. It is a top-level message field
// in pusher's shape, not part of the notification block.
func apsContentAvailable(payload map[string]interface{}) bool {
	switch v := payload["content-available"].(type) {
	case bool:
		return v
	case float64:
		return v == 1
	case int:
		return v == 1
	case string:
		return v == "1" || v == "true"
	}
	return false
}

func stringField(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

// joinLocArgs flattens APNs loc-args (a JSON array) into the single string
// pusher's Notification models (pusher re-wraps it into a one-element array).
// Multiple args are comma-joined — a known parity limitation of pusher's shape.
func joinLocArgs(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case []interface{}:
		parts := make([]string, 0, len(t))
		for _, e := range t {
			parts = append(parts, fmt.Sprintf("%v", e))
		}
		return strings.Join(parts, ",")
	}
	return ""
}

// numberToString renders an APNs badge (a JSON number) as the string pusher's
// Notification.Badge expects. JSON numbers decode to float64; integers are
// rendered without a trailing ".0".
func numberToString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	case int:
		return strconv.Itoa(t)
	case int64:
		return strconv.FormatInt(t, 10)
	case json.Number:
		return t.String()
	}
	return ""
}

// ToJSON returns the serialized message.
func (m *FCMMessage) ToJSON() (string, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
