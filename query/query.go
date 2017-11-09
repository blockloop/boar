package query

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
)

var errUseSlice = errors.New("array fields are not supported. use a slice instead")

// Parse parses query parameters from the http.Request and injects them into
// the struct v. Once v has been injected it runs govalidator against the struct
// and returns error if any validation fails
func Parse(v interface{}, r *http.Request) error {
	obj := reflect.ValueOf(v).Elem()
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
		if tag, ok := tField.Tag.Lookup("q"); ok {
			queryKey = tag
		}

		query := r.URL.Query()
		vals := query[queryKey]

		if len(vals) == 0 {
			continue
		}

		if kind == reflect.Slice {
			if err := setFieldSlice(field, tField.Name, vals); err != nil {
				return err
			}
			continue
		}

		val := vals[0]
		if val == "" {
			continue
		}
		err := setSimpleField(field, tField.Name, kind, val)
		if err != nil {
			return err
		}

	}
	return nil
}

func setFieldSlice(field reflect.Value, fieldName string, vals []string) error {
	len := len(vals)
	if len == 0 {
		return nil
	}
	fieldType := field.Type()
	for _, v := range vals {
		elemType := fieldType.Elem()
		fieldVal := reflect.New(elemType)
		if err := setSimpleField(reflect.Indirect(fieldVal), fieldName, elemType.Kind(), v); err != nil {
			return err
		}
		field.Set(reflect.Append(field, reflect.Indirect(fieldVal)))
	}
	return nil
}

func setSimpleField(f reflect.Value, fieldName string, kind reflect.Kind, val string) error {
	switch kind {
	case reflect.String:
		f.SetString(val)
		return nil
	case reflect.Bool:
		v, err := strconv.ParseBool(val)
		if err != nil {
			return &typeMismatchError{
				kind:      kind,
				val:       val,
				cause:     err,
				fieldName: fieldName,
			}
		}
		f.SetBool(v)
		break
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return &typeMismatchError{
				kind:      kind,
				val:       val,
				cause:     err,
				fieldName: fieldName,
			}
		}
		f.SetInt(v)
		break
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return &typeMismatchError{
				kind:      kind,
				val:       val,
				cause:     err,
				fieldName: fieldName,
			}
		}
		f.SetUint(v)
		break
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return &typeMismatchError{
				kind:      kind,
				val:       val,
				cause:     err,
				fieldName: fieldName,
			}
		}
		f.SetFloat(v)
		return nil
	default:
		return fmt.Errorf("%s is not a supported query parameter type", kind)
	}

	return nil
}

var _ error = (*typeMismatchError)(nil)

type typeMismatchError struct {
	kind      reflect.Kind
	val       interface{}
	cause     error
	fieldName string
}

func (e typeMismatchError) Error() string {
	return fmt.Sprintf("%q is not a valid %q for parameter %q", e.val, e.kind, e.fieldName)
}
