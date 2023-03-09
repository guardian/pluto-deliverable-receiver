package helpers

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"github.com/MicahParks/keyfunc"
	"time"
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

func extractPublicKey(certPem string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(certPem))

	//var cert *x509.Certificate
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		log.Print("ERROR helpers.extractPublicKey could not parse incoming certificate: ", err)
		return nil, err
	}
	rsaPublicKey := cert.PublicKey.(*rsa.PublicKey)
	return rsaPublicKey, nil
}

func LoadPublicKey(filePath string) (string, error) {
	f, openErr := os.Open(filePath)
	if openErr != nil {
		log.Printf("ERROR helpers.LoadPublicKey could not load cert from '%s': %s", filePath, openErr)
		return "", openErr
	}
	defer f.Close()

	data, readErr := ioutil.ReadAll(f)
	if readErr != nil {
		log.Printf("ERROR helpers.LoadPublicKey could not read in all data from '%s': %s", filePath, readErr)
		return "", readErr
	}
	return string(data), nil
}

func ValidateLogin(h *http.Request, config *Config) (string, error) {
	rawData, rawErr := extractAuth(h)
	if rawErr != nil {
		return "", rawErr
	}

	var token *jwt.Token

	if strings.HasPrefix(config.JWT.CertFile, "http") {
		options := keyfunc.Options{
				RefreshErrorHandler: func(err error) {
						log.Printf("ERROR There was an error with the jwt.Keyfunc\nError: %s", err.Error())
				},
				RefreshInterval:   time.Hour,
				RefreshRateLimit:  time.Minute * 5,
				RefreshTimeout:    time.Second * 10,
				RefreshUnknownKID: true,
		}

		jwks, err := keyfunc.Get(config.JWT.CertFile, options)
		if err != nil {
				log.Printf("ERROR Failed to create JWKS from resource at the given URL.\nError: %s", err.Error())
				return "", errors.New("Error loading JWKS from given URL")
		}

		if token, err = jwt.Parse(rawData, jwks.Keyfunc); err != nil {
				log.Printf("ERROR Failed to parse the JWT.\nError: %s", err.Error())
				return "", errors.New("Error parsing token using JWKS")
		}
	} else {
		publicCertData, loadErr := LoadPublicKey(config.JWT.CertFile)
		if loadErr != nil {
			return "", errors.New("Server setup problem, see logs")
		}

		var tokErr error

		token, tokErr = jwt.Parse(rawData, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return extractPublicKey(publicCertData)
		})

		if tokErr != nil {
			if validationErr, isValidationErr := tokErr.(*jwt.ValidationError); isValidationErr {
				log.Print("ERROR helpers.jwt.ValidateLogin could not validate: ", validationErr.Error())
				if (validationErr.Errors | jwt.ValidationErrorExpired) != 0 {
					return "", errors.New("Token is expired")
				}
			} else {
				log.Printf("ERROR helpers.jwt.ValidateLogin could not validate token '%s': %s", rawData, tokErr)
			}
			return "", errors.New("Internal validation error")
		}
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
