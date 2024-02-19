package helper

import (
	"fmt"
	"net/http"
)

func makeHttpErrorResponseBody(message string) string {
	return fmt.Sprintf("{\"message\":\"%s\"}", message)
}

func MakeHttpInternalServerErrorResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(makeHttpErrorResponseBody("Erro interno do servidor")))
}

func MakeHttpNotFoundErrorResponse(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(makeHttpErrorResponseBody(message)))
}

func MakeHttpUnprocessableEntityErrorResponse(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	w.Write([]byte(makeHttpErrorResponseBody(message)))
}

func MakeHttpResponseFromJSONString(w http.ResponseWriter, statusCode int, json string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write([]byte(json))
}

func MakeHttpResponseFromJSONBytes(w http.ResponseWriter, statusCode int, json []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(json)
}
