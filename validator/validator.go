// Package validator provides custom validation extensions.
package validator

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// CustomValidator wraps the go-playground validator.
type CustomValidator struct {
	Validator *validator.Validate
}

// New creates a new validator with common customizations.
func New() *CustomValidator {
	v := validator.New()

	// Use json tag name in error messages
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return fld.Name
		}
		return name
	})

	return &CustomValidator{Validator: v}
}

// Validate validates a struct.
func (cv *CustomValidator) Validate(i any) error {
	return cv.Validator.Struct(i)
}

// RegisterValidation registers a custom validation function.
func (cv *CustomValidator) RegisterValidation(tag string, fn validator.Func) error {
	return cv.Validator.RegisterValidation(tag, fn)
}
