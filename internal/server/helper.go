package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/SameerJadav/go-api/internal/logger"
)

func handleJSONDecodeErrors(w http.ResponseWriter, err error) {
	var syntaxError *json.SyntaxError
	var unmarshalTypeError *json.UnmarshalTypeError

	switch {
	// catch syntax error in JSON and send error message with location
	// of the problem to make it easier for client to fix
	case errors.As(err, &syntaxError):
		msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
		http.Error(w, msg, http.StatusBadRequest)

	// in some circumstances Decode() may also return an
	// io.ErrUnexpectedEOF error for syntax errors in the JSON
	case errors.Is(err, io.ErrUnexpectedEOF):
		msg := "Request body contains badly-formed JSON"
		http.Error(w, msg, http.StatusBadRequest)

	// catch any type errors, like trying to assign a string in the
	// request body to a int field in our User struct
	// send error message with field name and position of error
	case errors.As(err, &unmarshalTypeError):
		msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
		http.Error(w, msg, http.StatusBadRequest)

	// catch the error caused by extra unexpected fields in the request body
	case strings.HasPrefix(err.Error(), "json: unknown field"):
		fieldName := strings.TrimSpace(strings.TrimPrefix(err.Error(), "json: unknown field"))
		msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
		http.Error(w, msg, http.StatusBadRequest)

	// if body is empty
	case errors.Is(err, io.EOF):
		msg := "Request body must not be empty"
		http.Error(w, msg, http.StatusBadRequest)

	// catch the error caused by the request body being too large
	case errors.Is(err, &http.MaxBytesError{}):
		msg := "Request body must not be larger than 1MB"
		http.Error(w, msg, http.StatusRequestEntityTooLarge)

	// otherwise default to logging the error and sending a 500 Internal Server Error response
	default:
		logger.Error.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func validateSignleJSONObject(w http.ResponseWriter, r *http.Request) bool {
	dec := json.NewDecoder(r.Body)

	// if the body has only one JSON io.EOF error will be returned
	// so if we get anything else, we know that there is additional data in the body
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		msg := "Request body must only contain a single JSON object"
		http.Error(w, msg, http.StatusBadRequest)
		return false
	}
	return true
}

func validateContentType(w http.ResponseWriter, r *http.Request) bool {
	ct := r.Header.Get("Content-Type")
	if ct != "" {
		mediaType := strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
		if mediaType != "application/json" {
			msg := "Content-Type header is not application/json"
			http.Error(w, msg, http.StatusUnsupportedMediaType)
			return false
		}
	}
	return true
}

func parseIDFromPath(w http.ResponseWriter, r *http.Request) (int, error) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		logger.Error.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return 0, err
	}
	if id < 1 {
		msg := "User not found"
		http.Error(w, msg, http.StatusNotFound)
		return 0, err
	}
	return id, nil
}
