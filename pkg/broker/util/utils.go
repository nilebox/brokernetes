package util

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// WriteResponse will serialize 'object' to the HTTP ResponseWriter
// using the 'code' as the HTTP status code
func WriteResponse(w http.ResponseWriter, code int, object interface{}) {
	data, err := json.Marshal(object)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
}

// WriteEmptyResponse will return empty JSON object '{}" as a response
// using the 'code' as the HTTP status code
func WriteEmptyResponse(w http.ResponseWriter, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write([]byte("{}"))
}

// WriteResponse will serialize 'object' to the HTTP ResponseWriter
// using the 'code' as the HTTP status code
func WriteErrorResponse(w http.ResponseWriter, code int, err error) {
	type e struct {
		Description string `json:"description"`
	}
	data, err := json.Marshal(&e{
		Description: err.Error(),
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
}

// BodyToObject will convert the incoming HTTP request into the
// passed in 'object'
func BodyToObject(r *http.Request, object interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, object)
	if err != nil {
		return err
	}

	return nil
}
