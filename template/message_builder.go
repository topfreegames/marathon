package template

import (
	"encoding/json"
	"fmt"

	"git.topfreegames.com/topfreegames/marathon/messages"
	"github.com/uber-go/zap"
	"github.com/valyala/fasttemplate"
)

// Logger is the consumer logger
var Logger = zap.NewJSON(zap.WarnLevel)

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
			Logger.Error(
				"Error building message",
				zap.Error(err),
			)
			continue
		}
		outChan <- kafkaMsg
	}
}

// ReplaceTemplate replaces the template parameters from message with the
// values of the keys in params. The function returns the built string
func ReplaceTemplate(message string, params map[string]interface{}) (string, error) {
	t, errT := fasttemplate.NewTemplate(message, "{{", "}}")
	if errT != nil {
		Logger.Error(
			"Template Error",
			zap.Error(errT),
		)
		return "", errT
	}
	s := t.ExecuteString(params)
	return s, nil
}

// BuildApnsMsg builds an apns message
func BuildApnsMsg(request *messages.RequestMessage, content string) (string, error) {
	msg := newApnsMessage()
	msg.DeviceToken = request.Token
	msg.PushExpiry = request.PushExpiry
	msg.Payload.Aps = json.RawMessage(content)
	if len(request.Metadata) > 0 {
		msg.Payload.M = request.Metadata
	}

	b, err := json.Marshal(msg)
	if err != nil {
		Logger.Error(
			"Error building apns msg",
			zap.String("msg", fmt.Sprintf("%+v", msg)),
			zap.Error(err),
		)
		return "", err
	}
	strMsg := string(b)
	return strMsg, nil
}

// BuildGcmMsg builds a GCM message
func BuildGcmMsg(request *messages.RequestMessage, content string) (string, error) {
	msg := newGcmMessage()
	msg.To = request.Token
	msg.PushExpiry = request.PushExpiry
	err := json.Unmarshal([]byte(content), &msg.Data)
	if err != nil {
		Logger.Error(
			"Error building gcm msg",
			zap.String("msg", fmt.Sprintf("%+v", msg)),
			zap.Error(err),
		)
		return "", err
	}
	if len(request.Metadata) > 0 {
		msg.Data["m"] = request.Metadata
	}

	b, err := json.Marshal(msg)
	if err != nil {
		Logger.Error(
			"Error building gcm msg",
			zap.String("msg", fmt.Sprintf("%+v", msg)),
			zap.Error(err),
		)
		return "", err
	}
	strMsg := string(b)
	return strMsg, nil
}

// BuildMessage receives a RequestMessage with Message filled with the message
// to be built, if Params has any content. If the request was a template it
// should have already been fetched and placed into the Message field.
func BuildMessage(request *messages.RequestMessage) (*messages.KafkaMessage, error) {
	// Replace template

	content, rplTplerr := ReplaceTemplate(request.Message, request.Params)
	if rplTplerr != nil {
		Logger.Error(
			"Template replacing error",
			zap.Error(rplTplerr),
		)
		return nil, rplTplerr
	}

	topic := request.App

	var (
		message string
		msgErr  error
	)
	if request.Type == "apns" {
		message, msgErr = BuildApnsMsg(request, content)
		if msgErr != nil {
			return nil, msgErr
		}
	} else if request.Type == "gcm" {
		message, msgErr = BuildGcmMsg(request, content)
		if msgErr != nil {
			return nil, msgErr
		}
	} else {
		return nil, buildError{"Unknown request message type"}
	}
	if message == "" {
		return nil, buildError{"Failed to build KafkaMessage"}
	}

	return &messages.KafkaMessage{Message: message, Topic: topic}, nil
}
