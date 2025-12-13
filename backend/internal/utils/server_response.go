package utils

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
)

// readJSON read json from request body into data. It accepts a sinle JSON of 1MB max size value in the body
func ReadJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1048576 //maximum allowable bytes is 1MB

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(data)
	if err != nil {
		return err
	}

	err = dec.Decode(&struct{}{})

	if err != io.EOF {
		return errors.New("body must only have a single JSON value")
	}

	return nil
}

// writeJSON writes arbitrary data out as json
func WriteJSON(w http.ResponseWriter, status int, data any, headers ...http.Header) error {
	out, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}
	//add the headers if exists
	if len(headers) > 0 {
		for i, v := range headers[0] {
			w.Header()[i] = v
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(out)
	return nil
}

// badRequest sends a JSON response with the status http.StatusBadRequest, describing the error
func BadRequest(w http.ResponseWriter, err error) {
	var payload struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	payload.Error = true
	payload.Message = err.Error()
	_ = WriteJSON(w, http.StatusBadRequest, payload)
}

// NotFound sends a 404 JSON response with a standard structure.
func NotFound(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Resource not found"
	}

	resp := struct {
		Error   bool   `json:"error"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}{
		Error:   true,
		Status:  "not_found",
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	_ = json.NewEncoder(w).Encode(resp)
}

// ServerError sends a 500 JSON response with a standard structure.
func ServerError(w http.ResponseWriter, err error) {
	message := "Internal server error"
	if err != nil && err.Error() != "" {
		message = err.Error()
	}

	resp := struct {
		Error   bool   `json:"error"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}{
		Error:   true,
		Status:  "server_error",
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(resp)
}

// Unauthorized sends an HTTP 401 Unauthorized response to the client.
// It takes the http.ResponseWriter and an error object (which supplies the message).
func Unauthorized(w http.ResponseWriter, err error) {
	// Optional: Log the fact that an unauthorized attempt occurred
	// If you already log this in the middleware/handler, you might remove this.
	// log.Printf("Unauthorized access attempt: %v", err)

	// Define the structured error response
	resp := struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}{
		Error:   true,
		Message: err.Error(),
	}

	// Write the JSON response with HTTP 401 status code
	if err := WriteJSON(w, http.StatusUnauthorized, resp); err != nil {
		// If WriteJSON fails (e.g., network error), log the failure.
		log.Printf("Error writing unauthorized response: %v", err)
	}
}
