package utils

import (
	"encoding/json"
	"net/http"
)

func JsonResponse(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json := json.NewEncoder(w)
	json.SetIndent("", "  ")
	json.Encode(data)
}
