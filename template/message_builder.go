package template

import (
	"bytes"
	"encoding/json"
	"fmt"

	"git.topfreegames.com/topfreegames/marathon/messages"
	"github.com/cbroglie/mustache"
	"github.com/golang/glog"
)

type buildError struct {
	Message string
}

func (e buildError) Error() string {
	return fmt.Sprintf("%v", e.Message)
}

// MessageBuilder reads the messages from the inChan and generates messages to be sent
// to Kafka in the outChan
func MessageBuilder(inChan <-chan *messages.RequestMessage, outChan chan<- *messages.KafkaMessage) {
	for inReq := range inChan {
		kafkaMsg, err := BuildMessage(inReq)
		if err != nil {
			glog.Errorf("Error building message: %+v", err)
			continue
		}
		outChan <- kafkaMsg
	}
}

// replaceTemplate replaces the template parameters from message with the
// values of the keys in params. The function returns the built string
func replaceTemplate(message string, params *map[string]string) string {
	tmpl, tplErr := mustache.ParseString(message)
	if tplErr != nil {
		glog.Errorf("Template error")
	}
	var buf bytes.Buffer
	msg, rndrErr := tmpl.Render(*params, &buf)
	if rndrErr != nil {
		glog.Errorf("Render error")
	}
	return msg
}

func buildApnsMsg(request *messages.RequestMessage, content string) string {
	msg := newApnsMessage()
	msg.DeviceToken = request.Token
	msg.PushExpiry = request.PushExpiry
	msg.Payload.Aps = json.RawMessage(content)
	if len(request.Metadata) > 0 {
		msg.Payload.M = request.Metadata
	}

	b, err := json.Marshal(msg)
	if err != nil {
		glog.Errorf("Error building apns msg: %+v", err)
		return ""
	}
	strMsg := string(b)
	return strMsg
}

func buildGcmMsg(request *messages.RequestMessage, content string) string {
	msg := newGcmMessage()
	msg.To = request.Token
	msg.PushExpiry = request.PushExpiry
	err := json.Unmarshal([]byte(content), &msg.Data)
	if err != nil {
		glog.Errorf("Error building gcm msg: %+v", err)
		return ""
	}
	if len(request.Metadata) > 0 {
		msg.Data["m"] = request.Metadata
	}

	b, err := json.Marshal(msg)
	if err != nil {
		glog.Errorf("Error building gcm msg: %+v", err)
		return ""
	}
	strMsg := string(b)
	return strMsg
}

// BuildMessage receives a RequestMessage with Message filled with the message
// to be built, if Params has any content. If the request was a template it
// should have already been fetched and placed into the Message field.
func BuildMessage(request *messages.RequestMessage) (*messages.KafkaMessage, error) {
	// Replace template
	content := replaceTemplate(request.Message, &request.Params)
	topic := request.App

	var message string
	if request.Type == "apns" {
		message = buildApnsMsg(request, content)
	} else if request.Type == "gcm" {
		message = buildGcmMsg(request, content)
	} else {
		return nil, buildError{"Unknown request message type"}
	}
	if message == "" {
		return nil, buildError{"Failed to build KafkaMessage"}
	}

	return &messages.KafkaMessage{Message: message, Topic: topic}, nil
}
