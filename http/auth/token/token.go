// Derived from reqfleet/shibuya; modified for lib.
// Package token parses HTTP Bearer credentials without application-specific
// JWT, session, cookie, or CSRF policy.
package token

import (
	"errors"
	"net/http"
	"strings"
)

const (
	AuthHeader    = "Authorization"
	BEARER_PREFIX = "Bearer"
)

var (
	ErrEmptyToken   = errors.New("bearer header is empty")
	ErrInvalidToken = errors.New("token is invalid")
)

func ParseToken(bearer string) (string, error) {
	if bearer == "" {
		return "", ErrEmptyToken
	}
	prefix, token, ok := strings.Cut(bearer, " ")
	if !ok || prefix != BEARER_PREFIX || strings.Contains(token, " ") {
		return "", ErrInvalidToken
	}
	return token, nil
}

func FindBearerToken(header http.Header) (string, error) {
	bearer := header.Get(AuthHeader)
	token, err := ParseToken(bearer)
	if err != nil {
		return "", err
	}
	return token, nil
}
