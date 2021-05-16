package main

import (
	"encoding/json"
	"net/http"
)

var (
	errNoData               = &Error{"no_data", 400, "No data found", "No data found"}
	errBadRequest           = &Error{"bad_request", 400, "Bad request", "Request body is not well-formed. It must be JSON."}
	errUnauthorised         = &Error{"unauthorised", 401, "Unauthorised", "Make sure you pass all required credentials to your request"}
	errNotAcceptable        = &Error{"not_acceptable", 406, "Not Acceptable", "Accept header must be set to 'application/json'."}
	errDuplicateResource    = &Error{"duplicate_data", 409, "Resource already exists", "Resource already exists"}
	errUnsupportedMediaType = &Error{"unsupported_media_type", 415, "Unsupported Media Type", "Content-Type header must be set to: 'application/json'."}
	errInternalServer       = &Error{"internal_server_error", 500, "Internal Server Error", "Something went wrong."}
)

type Errors struct {
	Errors []*Error `json:"errors"`
}
type Error struct {
	Code        string `json:"code"`
	StatusCode  int    `json:"status"`
	Info        string `json:"info"`
	Description string `json:"description"`
}

func writeError(w http.ResponseWriter, err *Error) {

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(err.StatusCode)

	json.NewEncoder(w).Encode(Errors{[]*Error{err}})
}

func writeJSON(w http.ResponseWriter, data interface{}, statusCode int) {

	js, err := json.Marshal(data)

	if err != nil {

		writeError(w, errInternalServer)

		return
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(statusCode)

	w.Write(js)

	/*
			you could do the below instead but there is a risk if the Encoder panics,
		    then you'll get 200 code alongside with the panic!
	*/
	// w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(http.StatusOK)
	// json.NewEncoder(w).Encode(j)
}

// https://www.reddit.com/r/golang/comments/7p35s4/how_do_i_get_the_response_status_for_my_middleware/
type extendedResponseWriter struct {
	http.ResponseWriter

	status int
	length int
}

func (w *extendedResponseWriter) WriteHeader(status int) {

	w.status = status

	w.ResponseWriter.WriteHeader(status)
}

func (w *extendedResponseWriter) Write(b []byte) (int, error) {

	if w.status == 0 {

		w.status = 200
	}

	n, err := w.ResponseWriter.Write(b)

	w.length += n

	return n, err
}
