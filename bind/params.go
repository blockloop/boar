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
	return ParamsValue(reflect.ValueOf(v).Elem(), params)
}

// ParamsValue parses httprouter.Params and injects them into v.
func ParamsValue(obj reflect.Value, params httprouter.Params) error {
	kind := obj.Type()

	for i := 0; i < obj.NumField(); i++ {
		field := obj.Field(i)
		if !field.CanSet() {
			continue
		}

		tField := kind.Field(i)

		kind := tField.Type.Kind()
		// switch is benchmarked as about 5x faster than using a slice
		switch kind {
		case reflect.Complex64, reflect.Complex128, reflect.Array, reflect.Chan,
			reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice,
			reflect.Struct, reflect.UnsafePointer:
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
