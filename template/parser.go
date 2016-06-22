package template

import (
	"encoding/json"
	"fmt"

	"git.topfreegames.com/topfreegames/marathon/messages"
)

type parseError struct {
	Message string
}

func (e parseError) Error() string {
	return fmt.Sprintf("%v", e.Message)
}

// Parser is the function responsible to process the messages received from
// kafka, it listens to the channel from kafka and writes to the channel
// listened by TemplateFetcher.
// Multiple Parser instances should be able to run in parallel.
func Parser(inChan <-chan string, outChan chan<- *messages.RequestMessage) {
	for msg := range inChan {
		req, err := Parse(msg)
		if err != nil {
			// Logger.Error("Error parsing message: %+v", err)
			continue
		}
		outChan <- req
	}
}

// Parse parses the message string received, it expects to be in JSON format,
// as defined by the RequestMessage struct
func Parse(message string) (*messages.RequestMessage, error) {
	msg := new(messages.RequestMessage)
	err := json.Unmarshal([]byte(message), msg)
	if err != nil {
		e := parseError{fmt.Sprintf("Error parsing JSON: %+v, the message was: %s", err, message)}
		return msg, e
	}

	e := parseError{}
	// All fields should be set, except by either Template & Params or Message
	if msg.App == "" || msg.Token == "" || msg.Type == "" {
		e = parseError{"One of the mandatory fields is missing"}
	}
	// Either Template & Params should be defined or Message should be defined
	// Not both at the same time
	if msg.Template != "" && msg.Params == nil {
		e = parseError{"Template defined, but not Params"}
	}
	if msg.Template != "" && msg.Message != "" {
		e = parseError{"Both Template and Message defined"}
	}
	if msg.Template == "" && msg.Message == "" {
		e = parseError{"Either Template or Message should be defined"}
	}
	if e.Message != "" {
		return &messages.RequestMessage{}, e
	}

	// Logger.Debug("Decoded message: %+v", *msg)
	return msg, nil
}
