package bind

import (
	"io"
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetFieldSliceShouldDoNothingWhenValsEmpty(t *testing.T) {
	var slice []string
	field := reflect.Indirect(reflect.ValueOf(&slice))
	err := setFieldSlice(field, "", make([]string, 0))
	assert.NoError(t, err)
}

func TestSetFieldSliceShouldCreateSlice(t *testing.T) {
	var slice []int
	field := reflect.Indirect(reflect.ValueOf(&slice))
	err := setFieldSlice(field, "", []string{"1", "2"})
	assert.NoError(t, err)
	assert.Len(t, slice, 2)
}

func TestSetFieldSliceShouldErrorOnComplexFieldTypes(t *testing.T) {
	var slice []func()
	field := reflect.Indirect(reflect.ValueOf(&slice))
	err := setFieldSlice(field, "", []string{"1", "2"})
	assert.Error(t, err)
}

func TestSetSimpleFieldSetsString(t *testing.T) {
	var item string
	field := reflect.Indirect(reflect.ValueOf(&item))
	err := setSimpleField(field, "myfield", reflect.String, "2")
	require.NoError(t, err)
	assert.Equal(t, "2", item)
}

func TestSetSimpleFieldSetsInts(t *testing.T) {
	kinds := []reflect.Kind{
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
	}
	types := []reflect.Type{
		reflect.TypeOf(int(10)),
		reflect.TypeOf(int8(10)),
		reflect.TypeOf(int16(10)),
		reflect.TypeOf(int32(10)),
		reflect.TypeOf(int64(10)),
	}

	for i, kind := range kinds {
		field := reflect.Indirect(reflect.New(types[i]))
		err := setSimpleField(field, "myfield", kind, "10")
		require.NoError(t, err)
		assert.EqualValues(t, 10, field.Interface())
	}
}

func TestSetSimpleFieldSetsUints(t *testing.T) {
	kinds := []reflect.Kind{
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
	}
	types := []reflect.Type{
		reflect.TypeOf(uint(10)),
		reflect.TypeOf(uint8(10)),
		reflect.TypeOf(uint16(10)),
		reflect.TypeOf(uint32(10)),
		reflect.TypeOf(uint64(10)),
	}

	for i, kind := range kinds {
		field := reflect.Indirect(reflect.New(types[i]))
		err := setSimpleField(field, "myfield", kind, "10")
		require.NoError(t, err)
		assert.EqualValues(t, 10, field.Interface())
	}
}

func TestSetSimpleFieldSetsFloats(t *testing.T) {
	kinds := []reflect.Kind{
		reflect.Float32,
		reflect.Float64,
	}
	types := []reflect.Type{
		reflect.TypeOf(float32(10.0)),
		reflect.TypeOf(float64(10.0)),
	}

	for i, kind := range kinds {
		field := reflect.Indirect(reflect.New(types[i]))
		err := setSimpleField(field, "myfield", kind, "10")
		require.NoError(t, err)
		assert.EqualValues(t, 10, field.Interface())
	}
}

func TestSetSimpleFieldSetsBools(t *testing.T) {
	vals := []string{"t", "T", "f", "F", "1", "0", "true", "false"}

	for _, val := range vals {
		var field bool
		err := setSimpleField(reflect.Indirect(reflect.ValueOf(&field)), "myfield", reflect.Bool, val)
		require.NoError(t, err)
		expected, _ := strconv.ParseBool(val)
		assert.EqualValues(t, expected, field)
	}
}

func TestTypeMismatchErrors(t *testing.T) {
	kinds := []reflect.Kind{
		reflect.Int,
		reflect.Bool,
		reflect.Uint,
		reflect.Float32,
	}
	types := []reflect.Type{
		reflect.TypeOf(10),
		reflect.TypeOf(true),
		reflect.TypeOf(uint(10)),
		reflect.TypeOf(float32(10)),
	}

	for i, kind := range kinds {
		field := reflect.Indirect(reflect.New(types[i]))
		err := setSimpleField(field, "myfield", kind, "abcd")
		assert.IsType(t, err, &TypeMismatchError{})
	}
}

func TestTypeMismatchErrorShouldExplain(t *testing.T) {
	tme := &TypeMismatchError{
		cause:     io.ErrClosedPipe,
		fieldName: "asdfaksjdfh",
		kind:      reflect.Int,
		val:       "klasdhjf",
	}

	str := tme.Error()
	assert.Contains(t, str, tme.cause.Error())
	assert.Contains(t, str, tme.fieldName)
	assert.Contains(t, str, tme.kind.String())
	assert.Contains(t, str, tme.val)
}
