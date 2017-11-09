package bind

import (
	"fmt"
	"reflect"

	"github.com/julienschmidt/httprouter"
)

const (
	paramTagKey = "url"
)

// Params parses httprouter.Params and injects them into v.
func Params(v interface{}, params httprouter.Params) error {
	obj := reflect.ValueOf(v).Elem()
	kind := obj.Type()

	for i := 0; i < obj.NumField(); i++ {
		field := obj.Field(i)
		if !field.CanSet() {
			continue
		}

		tField := kind.Field(i)

		kind := tField.Type.Kind()
		switch kind {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Bool, reflect.Float32, reflect.Float64, reflect.String:
			break
		default:
			return fmt.Errorf("%q is not a supported type for a url parameter", kind)
		}

		queryKey := tField.Name
		if tag, ok := tField.Tag.Lookup(paramTagKey); ok {
			queryKey = tag
		}

		if queryKey == "-" {
			continue
		}

		val := params.ByName(queryKey)
		if len(val) == 0 {
			continue
		}

		err := setSimpleField(field, tField.Name, kind, val)
		if err != nil {
			return err
		}

	}
	return nil
}
