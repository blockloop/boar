package boar

import (
	"net/url"
	"reflect"
	"testing"

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
