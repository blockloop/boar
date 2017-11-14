// package bind provides reflection shortcuts for binding key/value pairs and strings
// to static types
package bind

import (
	"fmt"
	"reflect"
	"strconv"
)

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
			return &TypeMismatchError{
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
		if err != nil || f.OverflowInt(v) {
			return &TypeMismatchError{
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
		if err != nil || f.OverflowUint(v) {
			return &TypeMismatchError{
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
		if err != nil || f.OverflowFloat(v) {
			return &TypeMismatchError{
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

var _ error = (*TypeMismatchError)(nil)

// TypeMismatchError is an error that is caused by attempting to bind a type
// to a field with a different type.
type TypeMismatchError struct {
	kind      reflect.Kind
	val       interface{}
	cause     error
	fieldName string
}

func (e TypeMismatchError) Error() string {
	return fmt.Sprintf("%s' is not a valid %s for %s: %s", e.val, e.kind, e.fieldName, e.cause)
}
