package web

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

type customStatusError struct {
	code int
	msg  string
}

func (e *customStatusError) Error() string {
	return e.msg
}

func (e *customStatusError) StatusCode() int {
	return e.code
}

func TestHTTPErrorHandler(t *testing.T) {
	e := echo.New()

	tests := []struct {
		name           string
		err            error
		committed      bool
		expectedCode   int
		expectedBody   string
		shouldBeCalled bool
	}{
		{
			name:         "echo.HTTPError with message",
			err:          echo.NewHTTPError(http.StatusBadRequest, "invalid request"),
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":400,"message":"invalid request"}`,
		},
		{
			name:         "echo.HTTPError without message",
			err:          echo.NewHTTPError(http.StatusNotFound, ""),
			expectedCode: http.StatusNotFound,
			expectedBody: `{"code":404,"message":"Not Found"}`,
		},
		{
			name:         "echo.HTTPStatusCoder",
			err:          &customStatusError{code: http.StatusTeapot, msg: "I'm a teapot"},
			expectedCode: http.StatusTeapot,
			expectedBody: `{"code":418,"error":"I'm a teapot","message":"I'm a teapot"}`,
		},
		{
			name:         "generic error",
			err:          errors.New("something went wrong"),
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"code":500,"error":"something went wrong","message":"Internal Server Error"}`,
		},
		{
			name:         "committed response",
			err:          errors.New("something went wrong"),
			committed:    true,
			expectedCode: http.StatusOK, // Default for httptest.NewRecorder()
			expectedBody: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if tt.committed {
				if r, ok := echo.UnwrapResponse(c.Response()); ok == nil {
					r.Committed = true
				}
			}

			HTTPErrorHandler(c, tt.err)

			if !tt.committed {
				assert.Equal(t, tt.expectedCode, rec.Code)
				assert.JSONEq(t, tt.expectedBody, rec.Body.String())
			} else {
				assert.Empty(t, rec.Body.String())
			}
		})
	}
}
