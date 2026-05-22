package handler

import (
	"encoding/json"
	"net/http"

	"github.com/dutakey/shellgate/internal/formatter"
)

func Models() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(formatter.ModelsResponse())
	}
}
