package helpers

import (
	"errors"
	"net/http"
	"regexp"
	"strings"
)

func ValidateLogin(h *http.Request, config *Config) (string, error) {
	authHeader := h.Header.Get("Authorization")
	if authHeader=="" {
		return "", errors.New("no authorization header provided")
	}
	if ! strings.HasPrefix(authHeader, "Token") {
		return "", errors.New("expecting token auth")
	}

	xtrctor := regexp.MustCompile()
}
