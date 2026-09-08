package authz

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPError(t *testing.T) {
	t.Run("unauthorized", func(t *testing.T) {
		code, body := HTTPError(ErrUnauthorized)
		assert.Equal(t, http.StatusUnauthorized, code)
		assert.Equal(t, errorBody{Code: http.StatusUnauthorized, Msg: "unauthorized"}, body)
	})

	t.Run("forbidden", func(t *testing.T) {
		code, body := HTTPError(ErrForbidden)
		assert.Equal(t, http.StatusForbidden, code)
		assert.Equal(t, errorBody{Code: http.StatusForbidden, Msg: "forbidden"}, body)
	})

	t.Run("typed not found", func(t *testing.T) {
		code, body := HTTPError(NotFound("tenant"))
		assert.Equal(t, http.StatusNotFound, code)
		assert.Equal(t, errorBody{Code: http.StatusNotFound, Msg: "tenant not found"}, body)
	})

	t.Run("message not found", func(t *testing.T) {
		code, body := HTTPError(errors.New("product not found"))
		assert.Equal(t, http.StatusNotFound, code)
		assert.Equal(t, errorBody{Code: http.StatusNotFound, Msg: "product not found"}, body)
	})

	t.Run("other errors stay bad request", func(t *testing.T) {
		code, body := HTTPError(errors.New("tenantName is required"))
		assert.Equal(t, http.StatusBadRequest, code)
		assert.Equal(t, errorBody{Code: http.StatusBadRequest, Msg: "tenantName is required"}, body)
	})
}
