package asc

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrNotFound     = errors.New("resource not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrBadRequest   = errors.New("bad request")
)

// APIError represents a parsed App Store Connect error response.
type APIError struct {
	Code   string
	Title  string
	Detail string
}

func (e *APIError) Error() string {
	switch {
	case e.Title != "" && e.Detail != "":
		return fmt.Sprintf("%s: %s", e.Title, e.Detail)
	case e.Title != "":
		return e.Title
	case e.Detail != "":
		return e.Detail
	case e.Code != "":
		return e.Code
	default:
		return "API error"
	}
}

func (e *APIError) Is(target error) bool {
	switch target {
	case ErrNotFound:
		return strings.EqualFold(e.Code, "NOT_FOUND")
	case ErrUnauthorized:
		return strings.EqualFold(e.Code, "UNAUTHORIZED")
	case ErrForbidden:
		return strings.EqualFold(e.Code, "FORBIDDEN")
	case ErrBadRequest:
		return strings.EqualFold(e.Code, "BAD_REQUEST")
	default:
		return false
	}
}
