package boar

import (
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckFieldShouldReturnNoErrorWhenFieldDoesNotExist(t *testing.T) {
	_, err := checkField(reflect.Value{}, "")
	assert.NoError(t, err)
}

func TestCheckFieldShouldReturnFalseWhenFieldDoesNotExist(t *testing.T) {
	ok, _ := checkField(reflect.Value{}, "")
	assert.False(t, ok)
}

func TestCheckFieldShouldErrorWhenFieldIsNotAStruct(t *testing.T) {
	var handler struct {
		Query int
	}
	_, err := checkField(reflect.ValueOf(handler).FieldByName(QueryField), "")
	assert.Error(t, err)
}

func TestCheckFieldErrorWhenFieldIsNotAStructShouldExplain(t *testing.T) {
	var handler struct {
		Query int
	}
	_, err := checkField(reflect.ValueOf(handler).FieldByName(QueryField), "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "struct")
}

func TestCheckFieldShouldErrorWhenFieldIsNotSettable(t *testing.T) {
	var handler struct {
		Query struct {
			Age string
		}
	}
	_, err := checkField(reflect.ValueOf(handler).FieldByName("Query"), "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "setable")
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

func MakeHandlerShouldCallErrorHandlerWhenNilHandler(t *testing.T) {
	var called bool

	r := NewRouter()
	r.SetErrorHandler(func(c Context, err error) {
		called = true
	})

	hndlr := r.makeHandler("GET", "/", func(Context) (Handler, error) {
		return nil, nil
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	hndlr(w, req, nil)
	assert.True(t, called)
}

func MakeHandlerShouldCallErrorHandlerWhenErrorOnCreateHandler(t *testing.T) {
	var called bool

	r := NewRouter()
	hErr := errors.New("")

	r.SetErrorHandler(func(c Context, err error) {
		called = true
		assert.Equal(t, hErr, err)
	})

	hndlr := r.makeHandler("GET", "/", func(Context) (Handler, error) {
		return nil, hErr
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	hndlr(w, req, nil)
	assert.True(t, called)
}
