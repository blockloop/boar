package bind

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
	err := Query(&qp, r.URL.Query())
	require.NoError(t, err)
	require.Equal(t, "brett", qp.Name)
}

func TestParseShouldLookupUsingTag(t *testing.T) {
	type QueryParams struct {
		Name string `query:"name"`
	}

	var qp QueryParams
	r := httptest.NewRequest(http.MethodGet, "/?name=brett", nil)
	err := Query(&qp, r.URL.Query())
	require.NoError(t, err)
	require.Equal(t, "brett", qp.Name)
}

func TestParseShouldIgnorePrivateFields(t *testing.T) {
	type QueryParams struct {
		name string
	}

	var qp QueryParams
	r := httptest.NewRequest(http.MethodGet, "/?name=brett", nil)
	err := Query(&qp, r.URL.Query())
	require.NoError(t, err)
	require.Equal(t, "", qp.name)
}

func TestParseShouldSetIntegers(t *testing.T) {
	type QueryParams struct {
		Age int
	}

	var qp QueryParams
	r := httptest.NewRequest(http.MethodGet, "/?Age=1", nil)
	err := Query(&qp, r.URL.Query())
	require.NoError(t, err)
	require.Equal(t, 1, qp.Age)
}

func TestParseShouldSetBools(t *testing.T) {
	type QueryParams struct {
		Happy bool
	}

	var qp QueryParams
	r := httptest.NewRequest(http.MethodGet, "/?Happy=1", nil)
	err := Query(&qp, r.URL.Query())
	require.NoError(t, err)
	require.True(t, qp.Happy)
}

func TestParseShouldSetFloats(t *testing.T) {
	type QueryParams struct {
		Num float64
	}

	var qp QueryParams
	r := httptest.NewRequest(http.MethodGet, "/?Num=1.1", nil)
	err := Query(&qp, r.URL.Query())
	require.NoError(t, err)
	require.Equal(t, 1.1, qp.Num)
}

func TestParseShouldSetSliceInt(t *testing.T) {
	type QueryParams struct {
		Nums []int
	}

	var qp QueryParams
	r := httptest.NewRequest(http.MethodGet, "/?Nums=1&Nums=2&Nums=3", nil)
	err := Query(&qp, r.URL.Query())
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
	err := Query(&qp, r.URL.Query())
	require.NoError(t, err)
	require.Len(t, qp.Nums, 2)

	assert.Contains(t, qp.Nums, true)
	assert.Contains(t, qp.Nums, false)
}

func TestParseShouldIgnoreDashTagValues(t *testing.T) {
	type QueryParams struct {
		Num int `query:"-"`
	}

	qp := &QueryParams{
		Num: 100,
	}

	r := httptest.NewRequest(http.MethodGet, "/?Num=asdf", nil)
	err := Query(qp, r.URL.Query())
	require.NoError(t, err)
	assert.Equal(t, qp.Num, 100)
}

func TestParseShouldErrForArrays(t *testing.T) {
	var QueryParams struct {
		Num [2]int `query:"-"`
	}

	r := httptest.NewRequest(http.MethodGet, "/?Num=1&Num=2", nil)
	err := Query(&QueryParams, r.URL.Query())
	assert.Equal(t, errUseSlice, err)
}

func TestParseSkipsFieldsWithNoQueryValues(t *testing.T) {
	type QueryParams struct {
		Num int `query:"num"`
	}

	qp := &QueryParams{Num: 10}

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, Query(qp, r.URL.Query()))
	assert.Equal(t, 10, qp.Num)
}

func TestParseSkipsFieldsWithEmptyQueryValues(t *testing.T) {
	type QueryParams struct {
		Num int `query:"num"`
	}

	qp := &QueryParams{Num: 10}

	r := httptest.NewRequest(http.MethodGet, "/?num=&name=hello", nil)
	err := Query(qp, r.URL.Query())
	assert.NoError(t, err)
}

func TestParseSkipsFieldsWithEmptySpacedQueryValues(t *testing.T) {
	type QueryParams struct {
		Num int `query:"num"`
	}

	qp := &QueryParams{Num: 10}

	r := httptest.NewRequest(http.MethodGet, "/?num=%20&name=hello", nil)
	err := Query(qp, r.URL.Query())
	assert.NoError(t, err)
}

func TestParseTrimsValues(t *testing.T) {
	var QueryParams struct {
		Num []int `query:"num"`
	}

	r := httptest.NewRequest(http.MethodGet, "/?num=+10&num=%2020%20&num=30+", nil)
	err := Query(&QueryParams, r.URL.Query())
	assert.NoError(t, err)
	assert.Contains(t, QueryParams.Num, 10)
	assert.Contains(t, QueryParams.Num, 20)
	assert.Contains(t, QueryParams.Num, 30)
}

func TestParseErrorsTypeMismatchForMultipleQueryValues(t *testing.T) {
	var QueryParams struct {
		Num int `query:"num"`
	}

	r := httptest.NewRequest(http.MethodGet, "/?num=1&num=2", nil)
	err := Query(&QueryParams, r.URL.Query())
	require.IsType(t, &TypeMismatchError{}, err)
	assert.Equal(t, err.(*TypeMismatchError).Cause, errMultiValueSimpleField)
}

func TestParseErrorsTypeMismatchForBadSimpleValues(t *testing.T) {
	var QueryParams struct {
		Num int `query:"num"`
	}

	r := httptest.NewRequest(http.MethodGet, "/?num=abcd", nil)
	err := Query(&QueryParams, r.URL.Query())
	assert.IsType(t, &TypeMismatchError{}, err)
}

func TestParseErrorsTypeMismatchForBadSliceValues(t *testing.T) {
	var QueryParams struct {
		Num []int `query:"num"`
	}

	r := httptest.NewRequest(http.MethodGet, "/?num=1&num=2&num=abcd", nil)
	err := Query(&QueryParams, r.URL.Query())
	assert.IsType(t, &TypeMismatchError{}, err)
}
