package boar

import (
	"fmt"
	"net/url"
	"reflect"

	"github.com/asaskevich/govalidator"
	"github.com/blockloop/boar/bind"
	"github.com/julienschmidt/httprouter"
)

const (
	queryField     = "Query"
	urlParamsField = "URLParams"
	bodyField      = "Body"
)

func checkField(field reflect.Value, handlerName string) (bool, error) {
	if !field.IsValid() {
		return false, nil
	}
	if field.Kind() != reflect.Struct {
		return false, fmt.Errorf("'%s' field of '%s' must be a struct", queryField, handlerName)
	}
	if !field.CanSet() {
		return false, fmt.Errorf("'%s' field of '%s' is not setable", queryField, handlerName)
	}
	return true, nil
}

func setQuery(handler reflect.Value, qs url.Values) error {
	field := handler.FieldByName(queryField)
	if ok, err := checkField(field, handler.Type().Name()); !ok {
		return err
	}
	if err := bind.QueryValue(field, qs); err != nil {
		return NewValidationError(queryField, err)
	}
	return validate(queryField, field.Addr().Interface())
}

func setURLParams(handler reflect.Value, params httprouter.Params) error {
	field := handler.FieldByName(urlParamsField)
	if ok, err := checkField(field, handler.Type().Name()); !ok {
		return err
	}
	if err := bind.ParamsValue(field, params); err != nil {
		return NewValidationError(urlParamsField, err)
	}
	return validate(urlParamsField, field.Addr().Interface())
}

func setBody(handler reflect.Value, c Context) error {
	field := handler.FieldByName(bodyField)
	if ok, err := checkField(field, handler.Type().Name()); !ok {
		return err
	}
	if err := c.ReadJSON(field.Addr().Interface()); err != nil {
		return NewValidationError(bodyField, err)
	}
	return validate(bodyField, field.Addr().Interface())
}

func validate(fieldName string, v interface{}) error {
	valid, err := govalidator.ValidateStruct(v)
	if valid {
		return nil
	}
	if verr, ok := err.(govalidator.Errors); ok {
		return NewValidationErrors(fieldName, verr.Errors())
	}
	if verr, ok := err.(govalidator.Error); ok {
		return NewValidationErrors(fieldName, []error{verr})
	}
	return err
}
