package template

import (
	"encoding/json"
)

// ApnsMessage might need to update the json encoding if we change to snake case
type ApnsMessage struct {
	DeviceToken string         `json:"DeviceToken"`
	Payload     PayloadContent `json:"Payload"`
	PushExpiry  int64          `json:"push_expiry"`
}

// PayloadContent stores payload content of apns message
type PayloadContent struct {
	Aps json.RawMessage        `json:"aps"`
	M   map[string]interface{} `json:"m,omitempty"`
}

func newApnsMessage() *ApnsMessage {
	msg := new(ApnsMessage)
	msg.Payload = PayloadContent{}
	return msg
}

// GcmMessage is the struct to store a gcm message
type GcmMessage struct {
	To         string                 `json:"to"`
	Data       map[string]interface{} `json:"data"`
	PushExpiry int64                  `json:"push_expiry"`
}

func newGcmMessage() *GcmMessage {
	msg := new(GcmMessage)
	msg.Data = make(map[string]interface{})
	return msg
}
