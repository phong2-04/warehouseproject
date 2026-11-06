package handler

import (
	"encoding/json"
	"errors"
	"net/http"
)

// decodeJSONBody doc JSON voi gioi han dung luong request body.
func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst any, maxBytes int64, disallowUnknown bool) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	decoder := json.NewDecoder(r.Body)
	if disallowUnknown {
		decoder.DisallowUnknownFields()
	}

	return decoder.Decode(dst)
}

// isRequestBodyTooLarge kiem tra request body vuot gioi han.
func isRequestBodyTooLarge(err error) bool {
	var maxBytesErr *http.MaxBytesError
	return errors.As(err, &maxBytesErr)
}
