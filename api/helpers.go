package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"runtime"
	"strings"
	"unicode"
	"unicode/utf8"

	"git.topfreegames.com/topfreegames/marathon/log"
	"github.com/labstack/echo"
	"github.com/uber-go/zap"
)

// GetFileLine a file line
func GetFileLine() int {
	_, _, fileLine, _ := runtime.Caller(1)
	return fileLine
}

// GetFileName a file line
func GetFileName() string {
	_, fileName, _, _ := runtime.Caller(1)
	return fileName
}

// FailWith fails with the specified message
func FailWith(status int, message string, c echo.Context) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	return c.String(status, fmt.Sprintf(`{"success":false,"reason":"%s"}`, message))
}

// SucceedWith sends payload to user with status 200
func SucceedWith(payload map[string]interface{}, c echo.Context) error {
	if len(payload) == 0 {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		return c.String(200, `{"success":true}`)
	}
	payload["success"] = true
	return c.JSON(200, payload)
}

//LoadJSONPayload loads the JSON payload to the given struct validating all fields are not null
func LoadJSONPayload(payloadStruct interface{}, c echo.Context, l zap.Logger) error {
	log.D(l, "Loading payload...")

	data, err := GetRequestBody(c)
	if err != nil {
		log.E(l, "Loading payload failed.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return err
	}

	err = json.Unmarshal([]byte(data), payloadStruct)
	if err != nil {
		log.E(l, "Loading payload failed.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return err
	}

	var jsonPayload map[string]interface{}
	err = json.Unmarshal([]byte(data), &jsonPayload)
	if err != nil {
		log.E(l, "Loading payload failed.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return err
	}

	var missingFieldErrors []string
	v := reflect.ValueOf(payloadStruct).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		r, n := utf8.DecodeRuneInString(t.Field(i).Name)
		field := string(unicode.ToLower(r)) + t.Field(i).Name[n:]
		if jsonPayload[field] == nil {
			missingFieldErrors = append(missingFieldErrors, fmt.Sprintf("%s is required", field))
		}
	}

	if len(missingFieldErrors) != 0 {
		err := errors.New(strings.Join(missingFieldErrors[:], ", "))
		log.E(l, "Loading payload failed.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return err
	}

	log.D(l, "Payload loaded successfully.")
	return nil
}

//GetRequestBody from echo context
func GetRequestBody(c echo.Context) (string, error) {
	bodyCache := c.Get("requestBody")
	if bodyCache != nil {
		return bodyCache.(string), nil
	}
	body := c.Request().Body()
	b, err := ioutil.ReadAll(body)
	if err != nil {
		return "", err
	}
	c.Set("requestBody", string(b))
	return string(b), nil
}

//GetRequestJSON as the specified interface from echo context
func GetRequestJSON(payloadStruct interface{}, c echo.Context) error {
	body, err := GetRequestBody(c)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(body), payloadStruct)
	if err != nil {
		return err
	}

	return nil
}
