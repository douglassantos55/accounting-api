package api

import (
	"fmt"
	"strings"

	"example.com/accounting/database"
	"github.com/go-playground/validator/v10"
)

var databaseUnique validator.Func = func(fl validator.FieldLevel) bool {
	db, err := database.GetConnection()
	if err != nil {
		return false
	}

	field := fl.FieldName()
	entity := fl.Parent().Interface()

	result := map[string]interface{}{}
	tx := db.Model(entity).Where(entity, field).Not(entity, "ID")

	if result := tx.First(&result); result.RowsAffected != 0 {
		return false
	}

	return true
}

var validCpfCpnj validator.Func = func(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return IsCPF(value) || IsCNPJ(value)
}

type MyError struct {
	err validator.FieldError
}

func (e MyError) Field() string {
	_, name, _ := strings.Cut(e.err.Namespace(), ".")
	name = strings.Replace(name, "[", ".", -1)
	return strings.Replace(name, "]", "", -1)
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
