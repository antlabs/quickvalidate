package errors

import (
	"bytes"
	"fmt"
	"reflect"
)

// ValidationErrors is a collection of ValidationError
type ValidationErrors []ValidationError

// Error returns a string representation of the ValidationErrors
func (ve ValidationErrors) Error() string {
	buff := bytes.NewBufferString("")
	
	for i, err := range ve {
		if i > 0 {
			buff.WriteString(", ")
		}
		
		buff.WriteString(err.Error())
	}
	
	return buff.String()
}

// ValidationError stores information about a validation error
type ValidationError struct {
	Field     string
	Value     interface{}
	Tag       string
	Param     string
	Namespace string
}

// Error returns a string representation of the ValidationError
func (e ValidationError) Error() string {
	return fmt.Sprintf("Field validation for '%s' failed on the '%s' tag with value '%v'", 
		e.Field, e.Tag, e.Value)
}

// GetField returns the field name
func (e ValidationError) GetField() string {
	return e.Field
}

// GetValue returns the field value
func (e ValidationError) GetValue() interface{} {
	return e.Value
}

// GetTag returns the validation tag
func (e ValidationError) GetTag() string {
	return e.Tag
}

// GetParam returns the validation parameter
func (e ValidationError) GetParam() string {
	return e.Param
}

// GetNamespace returns the namespace
func (e ValidationError) GetNamespace() string {
	return e.Namespace
}

// InvalidValidationError is used when there is an error with the validation input
type InvalidValidationError struct {
	Type reflect.Type
}

// Error returns a string representation of the InvalidValidationError
func (e *InvalidValidationError) Error() string {
	if e.Type == nil {
		return "quickvalidate: Invalid validation input"
	}
	
	return fmt.Sprintf("quickvalidate: Invalid validation input type %s", e.Type)
}
