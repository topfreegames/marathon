package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/kataras/iris"
	"github.com/satori/go.uuid"
)

// FailWith fails with the specified message
func FailWith(status int, message string, c *iris.Context) {
	result, _ := json.Marshal(map[string]interface{}{
		"success": false,
		"reason":  message,
	})
	c.SetStatusCode(status)
	c.Write(string(result))
}

// SucceedWith sends payload to user with status 200
func SucceedWith(payload map[string]interface{}, c *iris.Context) {
	payload["success"] = true
	result, _ := json.Marshal(payload)
	c.SetStatusCode(200)
	c.Write(string(result))
}

// GetAsInt get a payload field as Int
func GetAsInt(field string, payload interface{}) int {
	v := reflect.ValueOf(payload)
	fieldValue := v.FieldByName(field).Interface()
	return fieldValue.(int)
}

// GetAsInt64 get a payload field as Int64
func GetAsInt64(field string, payload interface{}) int64 {
	v := reflect.ValueOf(payload)
	fieldValue := v.FieldByName(field).Interface()
	return fieldValue.(int64)
}

// GetAsJSON get a payload field as JSON
func GetAsJSON(field string, payload interface{}) map[string]interface{} {
	v := reflect.ValueOf(payload)
	fieldValue := v.FieldByName(field).Interface()
	return fieldValue.(map[string]interface{})
}

// GetAsArray get a payload field as Array
func GetAsArray(field string, payload interface{}) []string {
	v := reflect.ValueOf(payload)
	fieldValue := v.FieldByName(field).Interface()
	return fieldValue.([]string)
}

// GetAsString get a payload field as String
func GetAsString(field string, payload interface{}) string {
	v := reflect.ValueOf(payload)
	fieldValue := v.FieldByName(field).Interface()
	return fieldValue.(string)
}

// GetAsUUID get a payload field as UUID
func GetAsUUID(field string, payload interface{}) (uuid.UUID, error) {
	v := reflect.ValueOf(payload)
	fieldValue := v.FieldByName(field).Interface()
	return uuid.FromString(fieldValue.(string))
}

// GetAsRFC3339 get a payload field as RFC3339 time
func GetAsRFC3339(field string, payload interface{}) (time.Time, error) {
	v := reflect.ValueOf(payload)
	fieldValue := v.FieldByName(field).Interface()
	return time.Parse(time.RFC3339, fieldValue.(string))
}

// LoadJSONPayload loads the JSON payload to the given struct validating all fields are not null
func LoadJSONPayload(payloadStruct interface{}, c *iris.Context) (map[string]interface{}, error) {
	if err := c.ReadJSON(payloadStruct); err != nil {
		if err != nil {
			return nil, err
		}
	}

	data := c.RequestCtx.Request.Body()
	var jsonPayload map[string]interface{}
	err := json.Unmarshal(data, &jsonPayload)
	if err != nil {
		return nil, err
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
		error := errors.New(strings.Join(missingFieldErrors[:], ", "))
		return nil, error
	}

	return jsonPayload, nil
}
