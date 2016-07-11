package util

import (
	"encoding/json"
	"errors"

	"gopkg.in/gorp.v1"
)

//TypeConverter type
type TypeConverter struct{}

// ToDb converts val from json to string
func (tc TypeConverter) ToDb(val interface{}) (interface{}, error) {
	switch val.(type) {
	case map[string]interface{}:
		return json.Marshal(val)
	}
	return val, nil
}

// FromDb converts target from string to json
func (tc TypeConverter) FromDb(target interface{}) (gorp.CustomScanner, bool) {
	switch target.(type) {
	case *map[string]interface{}:
		binder := func(holder, target interface{}) error {
			s, ok := holder.(*string)
			if !ok {
				return errors.New("FromDb: Unable to convert map[string]interface{} to *string")
			}
			b := []byte(*s)
			return json.Unmarshal(b, target)
		}
		return gorp.CustomScanner{new(string), target, binder}, true
	}
	return gorp.CustomScanner{}, false
}
