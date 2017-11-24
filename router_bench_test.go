package boar

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	_ "net/http/pprof"
	"strconv"
	"testing"
)

type benchBodyHandler struct {
	handler HandlerFunc
	Body    struct {
		Name    string
		Age     int
		Charges []float32
	}
}

func (h *benchBodyHandler) Handle(c Context) error { return h.handler(c) }

func BenchmarkBoarHandlerWithBody(b *testing.B) {

	handler := &benchBodyHandler{}
	handler.handler = func(c Context) error {
		return nil
	}

	rtr := NewRouter()
	rtr.Get("/", func(Context) (Handler, error) {
		return handler, nil
	})

	rawReq := `{
		"Name": "Brett",
		"Age": 100,
		"Charges": [19.99, 20.99, 103.12]
	}`

	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(rawReq))
		rtr.ServeHTTP(rec, req)
		resp := rec.Result()

		if resp.Body != nil {
			resp.Body.Close()
		}
	}
}

func BenchmarkHTTPHandlerBaseWithBody(b *testing.B) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Name    string
			Age     int
			Charges []float32
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			b.Fatalf("could not decode body: %+v", err)
		}
		w.WriteHeader(http.StatusOK)
	})

	rawReq := `{
		"Name": "Brett",
		"Age": 100,
		"Charges": [19.99, 20.99, 103.12]
	}`

	for i := 0; i < b.N; i++ {
		r := httptest.NewRequest("POST", "/", bytes.NewBufferString(rawReq))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
	}
}

type benchBodyQueryHandler struct {
	handler HandlerFunc
	Body    struct {
		Name    string
		Age     int
		Charges []float32
	}
	Query struct {
		Page    int
		PerPage int
	}
}

func (h *benchBodyQueryHandler) Handle(c Context) error { return h.handler(c) }

func BenchmarkBoarHandlerWithBodyAndQuery(b *testing.B) {
	handler := &benchBodyHandler{
		handler: func(c Context) error {
			return nil
		},
	}

	rtr := NewRouter()
	rtr.Method("POST", "/", func(Context) (Handler, error) {
		return handler, nil
	})

	rawReq := `{
		"Name": "Brett",
		"Age": 100,
		"Charges": [19.99, 20.99, 103.12]
	}`

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/?Page=1&PerPage=100", bytes.NewBufferString(rawReq))
		req.Header.Set("content-type", contentTypeJSON)
		rec := httptest.NewRecorder()
		rtr.ServeHTTP(rec, req)
		resp := rec.Result()

		if resp.Body != nil {
			resp.Body.Close()
		}
	}
}

func BenchmarkHTTPHandlerBaseWithBodyAndQueryString(b *testing.B) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Name    string
			Age     int
			Charges []float32
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			b.Fatalf("could not decode body: %+v", err)
		}

		page, _ := strconv.ParseInt(r.URL.Query().Get("Page"), 10, 32)
		_ = int(page)

		perPage, _ := strconv.ParseInt(r.URL.Query().Get("PerPage"), 10, 32)
		_ = int(perPage)

		w.WriteHeader(http.StatusOK)
	}

	rawReq := `{
		"Name": "Brett",
		"Age": 100,
		"Charges": [19.99, 20.99, 103.12]
	}`

	for i := 0; i < b.N; i++ {
		r := httptest.NewRequest("POST", "/", bytes.NewBufferString(rawReq))
		w := httptest.NewRecorder()
		handler(w, r)
	}
}
