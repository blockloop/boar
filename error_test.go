package boar

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPErrorShouldPrintCause(t *testing.T) {
	err := &httpError{
		cause:  io.ErrClosedPipe,
		status: 100,
	}
	str := err.Error()
	assert.Contains(t, str, io.ErrClosedPipe.Error())
}

func TestHTTPErrorShouldPrintStatus(t *testing.T) {
	err := &httpError{
		cause:  io.ErrClosedPipe,
		status: 100,
	}
	str := err.Error()
	assert.Contains(t, str, "100")
}
