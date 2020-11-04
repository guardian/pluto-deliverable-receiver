package helpers

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"log"
	"net/http"
	"regexp"
	"strings"
)

func extractAuth(h *http.Request) (string, error) {
	authHeader := h.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("no authorization header provided")
	}
	if !strings.HasPrefix(authHeader, "Bearer") {
		return "", errors.New("expecting token auth")
	}

	xtrctor := regexp.MustCompile("^([^\\s]+)\\s+(.*)$")
	matches := xtrctor.FindAllStringSubmatch(authHeader, -1)

	if matches == nil {
		return "", errors.New("no token in auth header")
	}

	return matches[0][2], nil
}

func extractPublicKey(certPem string) *rsa.PublicKey {
	block, _ := pem.Decode([]byte(certPem))

	var cert *x509.Certificate
	cert, _ = x509.ParseCertificate(block.Bytes)
	rsaPublicKey := cert.PublicKey.(*rsa.PublicKey)
	return rsaPublicKey
}

func ValidateLogin(h *http.Request, config *Config) (string, error) {
	rawData, rawErr := extractAuth(h)
	if rawErr != nil {
		return "", rawErr
	}

	token, tokErr := jwt.Parse(rawData, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		key := extractPublicKey(config.JWT.PublicKeyPem)
		return key, nil
	})

	if tokErr != nil {
		if validationErr, isValidationErr := tokErr.(jwt.ValidationError); isValidationErr {
			log.Print("ERROR helpers.jwt.ValidateLogin could not validate: ", validationErr.Error())
		} else {
			log.Printf("ERROR helpers.jwt.ValidateLogin could not validate token '%s': %s", rawData, tokErr)
		}
		return "", errors.New("internal validation error")
	}

	if !token.Valid {
		log.Printf("ERROR token %s is not valid", rawData)
		return "", errors.New("token is not valid")
	}

	if claims, claimsOk := token.Claims.(jwt.MapClaims); claimsOk {
		for _, claimName := range config.JWT.UserNameClaims {
			if username, hasUsername := claims[claimName]; hasUsername {
				return username.(string), nil
			}
		}
		log.Printf("ERROR helpers.jwt.ValidateLogin token validated but could not get a username from any of %v", config.JWT.UserNameClaims)
		return "", errors.New("no username claim")
	}

	log.Printf("ERROR helpers.jwt.ValidateLogin claims data was not present or incorrect")
	return "", errors.New("incorrect claims data")
}
