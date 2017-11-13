// Code generated by mockery v1.0.0
package mocks

import http "net/http"
import mock "github.com/stretchr/testify/mock"

// ResponseWriter is an autogenerated mock type for the ResponseWriter type
type ResponseWriter struct {
	mock.Mock
}

// Body provides a mock function with given fields:
func (_m *ResponseWriter) Body() []byte {
	ret := _m.Called()

	var r0 []byte
	if rf, ok := ret.Get(0).(func() []byte); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	return r0
}

// Header provides a mock function with given fields:
func (_m *ResponseWriter) Header() http.Header {
	ret := _m.Called()

	var r0 http.Header
	if rf, ok := ret.Get(0).(func() http.Header); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Header)
		}
	}

	return r0
}

// Len provides a mock function with given fields:
func (_m *ResponseWriter) Len() int {
	ret := _m.Called()

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// Status provides a mock function with given fields:
func (_m *ResponseWriter) Status() int {
	ret := _m.Called()

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// Write provides a mock function with given fields: _a0
func (_m *ResponseWriter) Write(_a0 []byte) (int, error) {
	ret := _m.Called(_a0)

	var r0 int
	if rf, ok := ret.Get(0).(func([]byte) int); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]byte) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// WriteHeader provides a mock function with given fields: _a0
func (_m *ResponseWriter) WriteHeader(_a0 int) {
	_m.Called(_a0)
}
