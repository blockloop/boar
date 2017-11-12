package boar

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadJSONParsesJSONBody(t *testing.T) {
	body := bytes.NewBufferString(`{
		"name": "brett",
		"age": 100
	}`)
	r := httptest.NewRequest(http.MethodGet, "/", body)

	c := NewContext(r, nil, nil)

	type myStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	var req myStruct
	require.NoError(t, c.ReadJSON(&req), "read JSON")
	assert.NotNil(t, req)
	assert.Equal(t, "brett", req.Name)
	assert.Equal(t, 100, req.Age)
}

func TestReadJSONReturnsErrorIfJSONIsInvalid(t *testing.T) {
	body := bytes.NewBufferString(`}`)
	r := httptest.NewRequest(http.MethodGet, "/", body)

	c := NewContext(r, nil, nil)

	var req json.RawMessage
	err := c.ReadJSON(&req)
	assert.Error(t, err)
	assert.IsType(t, &httpError{}, err)
}

func TestReadJSONSetsBadRequestStatusIfJSONIsInvalid(t *testing.T) {
	body := bytes.NewBufferString(`}`)
	r := httptest.NewRequest(http.MethodGet, "/", body)

	c := NewContext(r, nil, nil)

	var req json.RawMessage
	err := c.ReadJSON(&req).(*httpError)
	assert.Equal(t, http.StatusBadRequest, err.status)
}
