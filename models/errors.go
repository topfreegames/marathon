package models

import "fmt"

// ModelNotFoundError identifies that a given model was not found in the Database with the given ID
type ModelNotFoundError struct {
	Type  string
	Field interface{}
	Value interface{}
}

func (e *ModelNotFoundError) Error() string {
	return fmt.Sprintf("%s was not found with %s: %v", e.Type, e.Field, e.Value)
}

// DuplicatedTokenError identifies that a given UserToken is duplicated in a given table
type DuplicatedTokenError struct {
	table string
	token string
}

func (e *DuplicatedTokenError) Error() string {
	return fmt.Sprintf("%s is duplicated in the table %s", e.token, e.table)
}
