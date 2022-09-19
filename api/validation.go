package api

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

var validCpfCpnj validator.Func = func(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return IsCPF(value) || IsCNPJ(value)
}

type MyError struct {
	err validator.FieldError
}

func (e MyError) Field() string {
	return e.err.Field()
}

func (e MyError) Error() string {
	return fmt.Sprintf("%s validation failed: %s", e.err.Field(), e.err.ActualTag())
}

func Errors(errors error) map[string]string {
	validationErrors := errors.(validator.ValidationErrors)
	out := map[string]string{}
	for _, err := range validationErrors {
		myerror := MyError{err}
		out[myerror.Field()] = myerror.Error()
	}
	return out
}
