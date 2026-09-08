package authz

import (
	"errors"
	"net/http"
	"strings"
)

type errorBody struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func HTTPError(err error) (int, any) {
	status := http.StatusBadRequest
	switch {
	case errors.Is(err, ErrUnauthorized):
		status = http.StatusUnauthorized
	case errors.Is(err, ErrForbidden):
		status = http.StatusForbidden
	case IsNotFound(err):
		status = http.StatusNotFound
	}
	return status, errorBody{Code: status, Msg: err.Error()}
}

func IsNotFound(err error) bool {
	var n *NotFoundError
	if errors.As(err, &n) {
		return true
	}
	msg := err.Error()
	return msg == "not found" || strings.HasSuffix(msg, " not found")
}
