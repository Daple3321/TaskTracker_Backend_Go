package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func WriteJSONResponse(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(v)
}

func ParseJSON(r *http.Request, payload any) error {
	if r.Body == nil {
		return fmt.Errorf("missing body request")
	}

	return json.NewDecoder(r.Body).Decode(payload)
}
