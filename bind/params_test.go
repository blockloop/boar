package bind

import (
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
)

func TestParamsParsesString(t *testing.T) {
	var item struct {
		Name string `url:"name"`
	}

	err := Params(&item, httprouter.Params{
		{Key: "name", Value: "Brett"},
	})
	assert.NoError(t, err)
	assert.Equal(t, item.Name, "Brett")
}

func TestParamsParsesInt(t *testing.T) {
	var item struct {
		Count int `url:"count"`
	}

	err := Params(&item, httprouter.Params{
		{Key: "count", Value: "1"},
	})
	assert.NoError(t, err)
	assert.Equal(t, item.Count, 1)
}

func TestParamsErrorsForSlice(t *testing.T) {
	var item struct {
		Count []int
	}

	err := Params(&item, httprouter.Params{
		{Key: "Count", Value: "[1]"},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "slice")
}

func TestParamsErrorsForArray(t *testing.T) {
	var item struct {
		Count [1]int
	}

	err := Params(&item, httprouter.Params{
		{Key: "Count", Value: "[1]"},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "array")
}
