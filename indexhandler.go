package main

import (
	"gitlab.com/codmill/customer-projects/guardian/deliverable-receiver/helpers"
	"net/http"
)

type IndexHandler struct {}

/**
print a very basic banner to show we are alive
 */
func (h IndexHandler) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	response := helpers.GenericErrorResponse{
		Status: "ok",
		Detail: "online",
	}

	helpers.WriteJsonContent(response, w, 200)
}