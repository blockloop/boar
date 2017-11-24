package bind

import (
	"errors"
	"net/url"
	"reflect"
	"strings"
)

var (
	errUseSlice              = errors.New("array fields are not supported. use a slice instead")
	errMultiValueSimpleField = errors.New("multiple values provided for non slice")
)

const (
	queryTagKey = "query"
)

// Query parses query parameters from the http.Request and injects them into v
func Query(v interface{}, q url.Values) error {
	return QueryValue(reflect.ValueOf(v).Elem(), q)
}

// QueryValue parses query parameters from the http.Request and injects them into v
func QueryValue(obj reflect.Value, q url.Values) error {
	kind := obj.Type()

	for i := 0; i < obj.NumField(); i++ {
		field := obj.Field(i)
		if !field.CanSet() {
			continue
		}

		tField := kind.Field(i)

		kind := tField.Type.Kind()
		if kind == reflect.Array {
			return errUseSlice
		}

		queryKey := tField.Name
		if tag, ok := tField.Tag.Lookup(queryTagKey); ok {
			queryKey = tag
		}

		if queryKey == "-" {
			continue
		}

		vals := q[queryKey]

		if len(vals) == 0 {
			continue
		}

		if kind == reflect.Slice {
			if err := setFieldSlice(field, tField.Name, vals); err != nil {
				return err
			}
			continue
		}

		// simple fields cannot have multiple values
		if len(vals) > 1 {
			return &TypeMismatchError{
				Cause:     errMultiValueSimpleField,
				FieldName: tField.Name,
				Kind:      kind,
				Val:       vals,
			}
		}

		val := strings.TrimSpace(vals[0])
		if val == "" {
			continue
		}
		err := setSimpleField(field, queryKey, kind, val)
		if err != nil {
			return err
		}

	}
	return nil
}
