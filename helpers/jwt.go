package helpers

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"regexp"
	"strings"
)

func extractAuth(h *http.Request) (string, error) {
	authHeader := h.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("no authorization header provided")
	}
	if !strings.HasPrefix(authHeader, "Token") {
		return "", errors.New("expecting token auth")
	}

	xtrctor := regexp.MustCompile("^([^\\s]+)\\s+(.*)$")
	matches := xtrctor.FindAllStringSubmatch(authHeader, -1)

	if matches == nil {
		return "", errors.New("no token in auth header")
	}

	return matches[0][2], nil
}

func ValidateLogin(h *http.Request, config *Config) (string, error) {
	rawData, rawErr := extractAuth(h)
	if rawErr != nil {
		return "", rawErr
	}

	token, tokErr := jwt.Parse(rawData, func(token *jwt.Token) (interface{}, error) {

	})
}
