package query

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseShouldLookupByFieldName(t *testing.T) {
	type QueryParams struct {
		Name string
	}

	var qp QueryParams
	r := httptest.NewRequest(http.MethodGet, "/?Name=brett", nil)
	err := Parse(&qp, r)
	require.NoError(t, err)
	require.Equal(t, "brett", qp.Name)
}

func TestParseShouldLookupUsingTag(t *testing.T) {
	type QueryParams struct {
		Name string `q:"name"`
	}

	var qp QueryParams
	r := httptest.NewRequest(http.MethodGet, "/?name=brett", nil)
	err := Parse(&qp, r)
	require.NoError(t, err)
	require.Equal(t, "brett", qp.Name)
}

func TestParseShouldIgnorePrivateFields(t *testing.T) {
	type QueryParams struct {
		name string
	}

	var qp QueryParams
	r := httptest.NewRequest(http.MethodGet, "/?name=brett", nil)
	err := Parse(&qp, r)
	require.NoError(t, err)
	require.Equal(t, "", qp.name)
}

func TestParseShouldSetIntegers(t *testing.T) {
	type QueryParams struct {
		Age int
	}

	var qp QueryParams
	r := httptest.NewRequest(http.MethodGet, "/?Age=1", nil)
	err := Parse(&qp, r)
	require.NoError(t, err)
	require.Equal(t, 1, qp.Age)
}

func TestParseShouldSetBools(t *testing.T) {
	type QueryParams struct {
		Happy bool
	}

	var qp QueryParams
	r := httptest.NewRequest(http.MethodGet, "/?Happy=1", nil)
	err := Parse(&qp, r)
	require.NoError(t, err)
	require.True(t, qp.Happy)
}

func TestParseShouldSetFloats(t *testing.T) {
	type QueryParams struct {
		Num float64
	}

	var qp QueryParams
	r := httptest.NewRequest(http.MethodGet, "/?Num=1.1", nil)
	err := Parse(&qp, r)
	require.NoError(t, err)
	require.Equal(t, 1.1, qp.Num)
}

func TestParseShouldSetSliceInt(t *testing.T) {
	type QueryParams struct {
		Nums []int
	}

	var qp QueryParams
	r := httptest.NewRequest(http.MethodGet, "/?Nums=1&Nums=2&Nums=3", nil)
	err := Parse(&qp, r)
	require.NoError(t, err)
	require.Len(t, qp.Nums, 3)
	assert.Contains(t, qp.Nums, 1)
	assert.Contains(t, qp.Nums, 2)
	assert.Contains(t, qp.Nums, 3)
}

func TestParseShouldSetSliceBool(t *testing.T) {
	type QueryParams struct {
		Nums []bool
	}

	var qp QueryParams
	r := httptest.NewRequest(http.MethodGet, "/?Nums=1&Nums=false", nil)
	err := Parse(&qp, r)
	require.NoError(t, err)
	require.Len(t, qp.Nums, 2)

	assert.Contains(t, qp.Nums, true)
	assert.Contains(t, qp.Nums, false)
}
