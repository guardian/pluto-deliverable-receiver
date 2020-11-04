package helpers

import (
	"net/http"
	"testing"
)

func TestExtractAuth(t *testing.T) {
	okRequest := &http.Request{
		Header: make(map[string][]string),
	}
	okRequest.Header.Add("Authorization", "Bearer sometoken")

	result, err := extractAuth(okRequest)
	if err != nil {
		t.Error("unexpected error from extractAuth: ", err)
	} else {
		if result != "sometoken" {
			t.Error("extractAuth should have returned 'sometoken', got '", result, "'")
		}
	}
}

//have not yet worked out how to do a _reliable_ test. this works, but only in the time window that the token is
//actually valid :(
/*
func TestValidateLogin(t *testing.T) {
	rawToken := ``
	signingCert := ``

	config := &Config{JWT: JwtConfig{
		PublicKeyPem:   signingCert,
		UserNameClaims: []string{"preferred_username"},
	}}

	okRequest := &http.Request{
		Header: make(map[string][]string),
	}
	okRequest.Header.Add("Authorization", "Bearer " + rawToken)

	userName, err := ValidateLogin(okRequest, config)
	if err != nil {
		t.Error("unexpected error from ValidateLogin: ", err)
	} else {
		if userName != "testuser" {
			t.Error("expected to get username 'testuser' from ValidateLogin, instead got '", userName, "'")
		}
	}
}
*/
