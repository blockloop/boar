package query_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/blockloop/boar/query"
)

func BenchmarkQueryParsingWithAllParams(b *testing.B) {
	var qp struct {
		Bools   []bool
		Nums    []int
		Name    string
		Age     int
		Money   float32
		Address string
		Debt    float32
	}

	qs := "?Nums=1&Nums=2&Nums=3&Nums=4&Nums=5&Nums=6&Nums=7&Nums=8"
	qs = qs + "&Bools=1&Bools=true&Bools=false&Bools=0&Bools=t&Bools=f"
	qs = qs + "&Name=brettjones&Age=99&Money=12.12&&Address=1999%ssomeRoad&Debg=999.99"
	r := httptest.NewRequest(http.MethodGet, "/"+qs, nil)

	for i := 0; i < b.N; i++ {
		err := query.Parse(&qp, r)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkQueryParsingWithAllParamsWithTags(b *testing.B) {
	var qp struct {
		Bools   []bool  `q:"bools"`
		Nums    []int   `q:"nums"`
		Name    string  `q:"name"`
		Age     int     `q:"age"`
		Money   float32 `q:"money"`
		Address string  `q:"address"`
		Debt    float32 `q:"debt"`
	}

	qs := "?nums=1&nums=2&nums=3&nums=4&nums=5&nums=6&nums=7&nums=8"
	qs = qs + "&bools=1&bools=true&bools=false&bools=0&bools=t&bools=f"
	qs = qs + "&name=brettjones&age=99&money=12.12&&address=1999%ssomeroad&debg=999.99"
	r := httptest.NewRequest(http.MethodGet, "/"+qs, nil)

	for i := 0; i < b.N; i++ {
		err := query.Parse(&qp, r)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkQueryParsingWithSimpleParams(b *testing.B) {
	var qp struct {
		Name    string
		Age     int
		Money   float32
		Address string
		Debt    float32
	}

	qs := "?Name=brettjones&Age=99&Money=12.12&&Address=1999%ssomeRoad&Debg=999.99"
	r := httptest.NewRequest(http.MethodGet, "/"+qs, nil)

	for i := 0; i < b.N; i++ {
		err := query.Parse(&qp, r)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkQueryParsingWithSimpleParamsWithTags(b *testing.B) {
	var qp struct {
		Name    string  `q:"Name"`
		Age     int     `q:"Age"`
		Money   float32 `q:"Money"`
		Address string  `q:"Address"`
		Debt    float32 `q:"Debt"`
	}

	qs := "?Name=brettjones&Age=99&Money=12.12&&Address=1999%ssomeRoad&Debg=999.99"
	r := httptest.NewRequest(http.MethodGet, "/"+qs, nil)

	for i := 0; i < b.N; i++ {
		err := query.Parse(&qp, r)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkQueryParsingWithOneSimpleParam(b *testing.B) {
	var qp struct {
		Name string
	}

	r := httptest.NewRequest(http.MethodGet, "/?Name=brett", nil)

	for i := 0; i < b.N; i++ {
		err := query.Parse(&qp, r)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkQueryParsingWithOneSliceParam(b *testing.B) {
	var qp struct {
		Names []string
	}

	r := httptest.NewRequest(http.MethodGet, "/?Names=brett&Names=kristy&Names=jack&Names=jill", nil)

	for i := 0; i < b.N; i++ {
		err := query.Parse(&qp, r)
		if err != nil {
			b.Fatal(err)
		}
	}
}
