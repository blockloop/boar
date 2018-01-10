package boar

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
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
			Name string `validate:"alpha"`
		}
	}
	err := setQuery(reflect.Indirect(reflect.ValueOf(&handler)), url.Values{
		"Name": []string{"1234"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Name")
}

func TestSetURLParamsReturnsNoErrorWhenFieldsAreOkay(t *testing.T) {
	var handler struct {
		URLParams struct {
			Age uint8
		}
	}

	params := httprouter.Params{
		{Key: "Age", Value: "40"},
	}
	err := setURLParams(reflect.Indirect(reflect.ValueOf(&handler)), params)
	assert.NoError(t, err)
}

func TestSetURLParamsShouldReturnNoErrorWhenFieldDoesNotExist(t *testing.T) {
	var handler struct{}
	err := setURLParams(reflect.Indirect(reflect.ValueOf(&handler)), nil)
	assert.NoError(t, err)
}

func TestSetURLParamsShouldReturnValidationErrorWhenValidationFails(t *testing.T) {
	var handler struct {
		URLParams struct {
			Name string `validate:"alpha"`
		}
	}
	key, badValue := "Name", "1234"
	err := setURLParams(reflect.Indirect(reflect.ValueOf(&handler)), httprouter.Params{
		{Key: key, Value: badValue},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), key)
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
}

func TestSetURLParamsReturnsValidationErrorWhenTypeMismatchErrorInBind(t *testing.T) {
	var handler struct {
		URLParams struct {
			Age int
		}
	}
	key, badValue := "Age", "abcd"
	err := setURLParams(reflect.Indirect(reflect.ValueOf(&handler)), httprouter.Params{
		{Key: key, Value: badValue},
	})
	assert.IsType(t, &ValidationError{}, err)
}

func TestSetURLParamsReturnsErrorWhenBadTypeForParameter(t *testing.T) {
	var handler struct {
		URLParams struct {
			Age func()
		}
	}
	key, badValue := "Age", "abcd"
	err := setURLParams(reflect.Indirect(reflect.ValueOf(&handler)), httprouter.Params{
		{Key: key, Value: badValue},
	})
	assert.Error(t, err)
	assert.Contains(t, fmt.Sprint(err), "not a supported type")
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mc := NewMockContext(ctrl)
	mc.EXPECT().ReadJSON(gomock.Any()).Return(expErr)
	req := httptest.NewRequest("GET", "/", bytes.NewBufferString("{}"))
	req.Header.Set("content-type", contentTypeJSON)
	mc.EXPECT().Request().Return(req)

	err := setBody(reflect.Indirect(reflect.ValueOf(&handler)), mc)
	if !assert.IsType(t, &ValidationError{}, err) {
		t.Error(err)
	}
}

func TestSetBodyShouldReturnValidationErrorWhenValidationFails(t *testing.T) {
	var handler struct {
		Body struct {
			Name string `validate:"alpha"`
		}
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mc := NewMockContext(ctrl)

	req := httptest.NewRequest("GET", "/", bytes.NewBufferString("{}"))
	req.Header.Set("content-type", contentTypeJSON)
	mc.EXPECT().Request().Return(req)

	mc.EXPECT().ReadJSON(gomock.Any()).Do(func(v interface{}) {
		json.Unmarshal([]byte(`{"Name": "1234"}`), v)
	}).Return(nil)

	err := setBody(reflect.Indirect(reflect.ValueOf(&handler)), mc)
	require.Error(t, err)
	assert.IsType(t, &ValidationError{}, err)
}

func TestSetBodyShouldReturnErrorWhenUnknownContentType(t *testing.T) {
	var handler struct {
		Body struct {
			Name string
		}
	}

	request := httptest.NewRequest("POST", "/", bytes.NewBufferString(`<xml></xml>`))
	request.Header.Set("content-type", "application/xml")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mc := NewMockContext(ctrl)
	mc.EXPECT().Request().Return(request)

	err := setBody(reflect.Indirect(reflect.ValueOf(&handler)), mc)
	require.Error(t, err)
}

func TestSetBodyUnknownContentTypeErrorShouldExplain(t *testing.T) {
	var handler struct {
		Body struct {
			Name string
		}
	}

	request := httptest.NewRequest("POST", "/", bytes.NewBufferString(`<xml></xml>`))
	request.Header.Set("content-type", "application/xml")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mc := NewMockContext(ctrl)
	mc.EXPECT().Request().Return(request)

	err := setBody(reflect.Indirect(reflect.ValueOf(&handler)), mc)
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "application/xml")
}

func TestSetBodyShouldRequireContentType(t *testing.T) {
	var handler struct {
		Body struct {
			Name string
		}
	}

	request := httptest.NewRequest("POST", "/", nil)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mc := NewMockContext(ctrl)
	mc.EXPECT().Request().Return(request)

	err := setBody(reflect.Indirect(reflect.ValueOf(&handler)), mc)
	assert.IsType(t, &httpError{}, err)
}

func TestSetBodyShouldParseFormForFormContentType(t *testing.T) {
	var handler struct {
		Body struct {
			Name string
		}
	}

	request := httptest.NewRequest("POST", "/", bytes.NewBufferString(`Name=brett`))
	request.Header.Set("content-type", contentTypeFormEncoded)
	w := httptest.NewRecorder()

	mc := NewContext(request, w, nil)

	err := setBody(reflect.Indirect(reflect.ValueOf(&handler)), mc)
	require.NoError(t, err)
	assert.Equal(t, "brett", handler.Body.Name)
}

func TestSetBodyShouldParseFormForMultipartFormContentType(t *testing.T) {
	var handler struct {
		Body struct {
			Name string
		}
	}

	rawReq := `POST / HTTP/1.1
Content-Length: 144
Expect: 100-continue
Content-Type: multipart/form-data; boundary=------------------------cee38e2aa6ef9de4

--------------------------cee38e2aa6ef9de4
Content-Disposition: form-data; name="Name"

brett
--------------------------cee38e2aa6ef9de4--
`

	request, err := http.ReadRequest(bufio.NewReader(bytes.NewBufferString(rawReq)))
	require.NoError(t, err)

	mc := NewContext(request, nil, nil)

	err = setBody(reflect.Indirect(reflect.ValueOf(&handler)), mc)
	require.NoError(t, err)
	assert.Equal(t, "brett", handler.Body.Name)
}

func TestValidateShouldErrorWhenBadValue(t *testing.T) {
	var f string

	err := validate("", &f)
	require.Error(t, err)
}
