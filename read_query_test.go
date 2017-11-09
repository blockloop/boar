package boar

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadQueryShouldReturnValidationRequestWithBadTypes(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?Age=hello", nil)

	c := NewContext(r, nil, nil)

	var q struct {
		Age int
	}

	err := c.ReadQuery(&q)
	assert.Error(t, err)
	assert.IsType(t, err, &ValidationError{})
}

func TestReadQueryShouldMentionCauseField(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?Age=hello", nil)

	c := NewContext(r, nil, nil)

	var q struct {
		Age int
	}

	err := c.ReadQuery(&q)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Age")
}

func TestReadQueryShouldMentionType(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?Age=hello", nil)

	c := NewContext(r, nil, nil)

	var q struct {
		Age int
	}

	err := c.ReadQuery(&q)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "int")
}

func TestReadQueryShouldSetBadRequestStatus(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?Age=hello", nil)

	c := NewContext(r, nil, nil)

	var q struct {
		Age int
	}

	err := c.ReadQuery(&q)
	assert.Equal(t, err.(*ValidationError).status, http.StatusBadRequest)
}
