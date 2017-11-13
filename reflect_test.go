package boar

import (
	"encoding/json"
	"errors"
	"net/url"
	"reflect"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	. "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCheckFieldShouldReturnNoErrorWhenFieldDoesNotExist(t *testing.T) {
	_, err := checkField(reflect.Value{})
	assert.NoError(t, err)
}

func TestCheckFieldShouldReturnFalseWhenFieldDoesNotExist(t *testing.T) {
	ok, _ := checkField(reflect.Value{})
	assert.False(t, ok)
}

func TestCheckFieldShouldErrorWhenFieldIsNotAStruct(t *testing.T) {
	var handler struct {
		Query int
	}
	_, err := checkField(reflect.ValueOf(handler).FieldByName(queryField))
	assert.Error(t, err)
}

func TestCheckFieldErrorWhenFieldIsNotAStructShouldExplain(t *testing.T) {
	var handler struct {
		Query int
	}
	_, err := checkField(reflect.ValueOf(handler).FieldByName(queryField))
	assert.Error(t, err)
	assert.Equal(t, err, errNotAStruct)
}

func TestCheckFieldShouldErrorWhenFieldIsNotSettable(t *testing.T) {
	var handler struct {
		Query struct {
			Age string
		}
	}
	_, err := checkField(reflect.ValueOf(handler).FieldByName(queryField))
	assert.Error(t, err)
	assert.Equal(t, err, errNotSettable)
}

func TestSetQueryShouldReturnNoErrorWhenFieldDoesNotExist(t *testing.T) {
	var handler struct{}
	err := setQuery(reflect.Indirect(reflect.ValueOf(&handler)), url.Values{})
	assert.NoError(t, err)
}

func TestSetQueryShouldErrorWithMismatchType(t *testing.T) {
	var handler struct {
		Query struct {
			Age int
		}
	}
	err := setQuery(reflect.Indirect(reflect.ValueOf(&handler)), url.Values{
		"Age": []string{"abcd"},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "abcd")
}

func TestSetQueryShouldErrorWhenValidationError(t *testing.T) {
	var handler struct {
		Query struct {
			Name string `valid:"alpha"`
		}
	}
	err := setQuery(reflect.Indirect(reflect.ValueOf(&handler)), url.Values{
		"Name": []string{"1234"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Name")
	assert.Contains(t, err.Error(), "1234")
}

func TestSetURLParamsShouldReturnNoErrorWhenFieldDoesNotExist(t *testing.T) {
	var handler struct{}
	err := setURLParams(reflect.Indirect(reflect.ValueOf(&handler)), nil)
	assert.NoError(t, err)
}

func TestSetURLParamsShouldReturnValidationErrorWhenValidationFails(t *testing.T) {
	var handler struct {
		URLParams struct {
			Name string `valid:"alpha"`
		}
	}
	key, badValue := "Name", "1234"
	err := setURLParams(reflect.Indirect(reflect.ValueOf(&handler)), httprouter.Params{
		{Key: key, Value: badValue},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), key)
	assert.Contains(t, err.Error(), badValue)
}

func TestSetURLParamsShouldReturnValidationErrorWhenBinderFails(t *testing.T) {
	var handler struct {
		URLParams struct {
			Age int
		}
	}
	key, badValue := "Age", "abcd"
	err := setURLParams(reflect.Indirect(reflect.ValueOf(&handler)), httprouter.Params{
		{Key: key, Value: badValue},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), key)
	assert.Contains(t, err.Error(), badValue)
}

func TestSetBodyShouldReturnNoErrorWhenFieldDoesNotExist(t *testing.T) {
	var handler struct{}
	err := setBody(reflect.Indirect(reflect.ValueOf(&handler)), nil)
	assert.NoError(t, err)
}

func TestSetBodyShouldReturnValidationErrorWhenCheckFieldFails(t *testing.T) {
	var handler struct {
		Body struct {
			Age int
		}
	}

	// handler is not a pointer and will fail checkField
	err := setBody(reflect.Indirect(reflect.ValueOf(handler)), nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), bodyField)
}

func TestSetBodyShouldReturnValidationErrorWhenReadJSONFails(t *testing.T) {
	var handler struct {
		Body struct {
			Age int
		}
	}

	expErr := errors.New("something went wrong")
	mc := &MockContext{}
	mc.On("ReadJSON", Anything).Return(expErr)

	err := setBody(reflect.Indirect(reflect.ValueOf(&handler)), mc)

	require.Error(t, err)
	assert.IsType(t, &ValidationError{}, err)
}

func TestSetBodyShouldReturnValidationErrorWhenValidationFails(t *testing.T) {
	var handler struct {
		Body struct {
			Name string `valid:"alpha"`
		}
	}

	mc := &MockContext{}
	mc.On("ReadJSON", Anything).Run(func(args Arguments) {
		json.Unmarshal([]byte(`{"Name": "1234"}`), args.Get(0))
	}).Return(nil)

	err := setBody(reflect.Indirect(reflect.ValueOf(&handler)), mc)
	require.Error(t, err)
	assert.IsType(t, &ValidationError{}, err)
}
