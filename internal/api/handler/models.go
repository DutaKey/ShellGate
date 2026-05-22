package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dutakey/shellgate/internal/types"
)

var availableModels = []types.Model{
	{ID: "gpt-5.5", Object: "model", Created: 1700000000, OwnedBy: "openai"},
	{ID: "gpt-5.4", Object: "model", Created: 1700000000, OwnedBy: "openai"},
	{ID: "gpt-5.4-mini", Object: "model", Created: 1700000000, OwnedBy: "openai"},
	{ID: "gpt-5.3-codex", Object: "model", Created: 1700000000, OwnedBy: "openai"},
	{ID: "gpt-5.2", Object: "model", Created: 1700000000, OwnedBy: "openai"},
}

func Models() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(types.ModelsResponse{
			Object: "list",
			Data:   availableModels,
		})
	}
}

func ModelByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		for _, m := range availableModels {
			if m.ID == id {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(m)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(types.ErrorResponse{
			Error: types.APIError{
				Message: "model '" + id + "' not found",
				Type:    "invalid_request_error",
				Code:    "model_not_found",
			},
		})
	}
}
