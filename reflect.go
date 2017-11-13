package boar

import (
	"errors"
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

var (
	errNotAStruct  = errors.New("not a struct")
	errNotSettable = errors.New("not settable")
)

func checkField(field reflect.Value) (bool, error) {
	if !field.IsValid() {
		return false, nil
	}
	if field.Kind() != reflect.Struct {
		return false, errNotAStruct
	}
	if !field.CanSet() {
		return false, errNotSettable
	}
	return true, nil
}

func setQuery(handler reflect.Value, qs url.Values) error {
	field := handler.FieldByName(queryField)
	ok, err := checkField(field)
	if !ok {
		if err == nil {
			return nil
		}
		return &badFieldError{
			field:   queryField,
			handler: handler,
			err:     err,
		}
	}
	if err := bind.QueryValue(field, qs); err != nil {
		return NewValidationError(queryField, err)
	}
	return validate(queryField, field.Addr().Interface())
}

func setURLParams(handler reflect.Value, params httprouter.Params) error {
	field := handler.FieldByName(urlParamsField)
	ok, err := checkField(field)
	if !ok {
		if err == nil {
			return nil
		}
		return &badFieldError{
			field:   urlParamsField,
			handler: handler,
			err:     err,
		}
	}
	if err := bind.ParamsValue(field, params); err != nil {
		return NewValidationError(urlParamsField, err)
	}
	return validate(urlParamsField, field.Addr().Interface())
}

func setBody(handler reflect.Value, c Context) error {
	field := handler.FieldByName(bodyField)
	ok, err := checkField(field)
	if !ok {
		if err == nil {
			return nil
		}
		return &badFieldError{
			field:   bodyField,
			handler: handler,
			err:     err,
		}
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

type badFieldError struct {
	field   string
	handler reflect.Value
	err     error
}

func (b badFieldError) Error() string {
	return fmt.Sprintf("%s field of %s is %s", b.field, b.handler.Type().Name(), b.err)
}
