package boar

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetQueryShouldDoNothingWhenQueryFieldDoesNotExist(t *testing.T) {
	var handler struct{}
	err := setQuery(reflect.ValueOf(handler), url.Values(nil))
	assert.NoError(t, err)
}

func TestSetQueryShouldErrorWhenQueryFieldIsNotAStruct(t *testing.T) {
	var handler struct {
		Query int
	}
	err := setQuery(reflect.ValueOf(handler), url.Values(nil))
	assert.Error(t, err)
}

func TestSetQueryErrorWhenQueryFieldIsNotAStructShouldExplain(t *testing.T) {
	var handler struct {
		Query int
	}
	err := setQuery(reflect.ValueOf(handler), url.Values(nil))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "struct")
}

func TestSetQueryShouldErrorWhenQueryFieldIsNotSettable(t *testing.T) {
	var handler struct {
		Query struct {
			Age string
		}
	}
	err := setQuery(reflect.ValueOf(handler), url.Values(nil))
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
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Name")
	assert.Contains(t, err.Error(), "1234")
}
