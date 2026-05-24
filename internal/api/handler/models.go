package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dutakey/shellgate/internal/types"
)

var codexModels = []types.Model{
	{ID: "gpt-5.5", Object: "model", Created: 1700000000, OwnedBy: "openai"},
	{ID: "gpt-5.4", Object: "model", Created: 1700000000, OwnedBy: "openai"},
	{ID: "gpt-5.4-mini", Object: "model", Created: 1700000000, OwnedBy: "openai"},
	{ID: "gpt-5.3-codex", Object: "model", Created: 1700000000, OwnedBy: "openai"},
	{ID: "gpt-5.2", Object: "model", Created: 1700000000, OwnedBy: "openai"},
}

var kimiModels = []types.Model{
	{ID: "kimi-code/kimi-for-coding", Object: "model", Created: 1700000000, OwnedBy: "moonshot-ai"},
}

// allModels merges models from all configured providers.
var allModels = append(append([]types.Model{}, codexModels...), kimiModels...)

// NewModelsHandlers returns (list, byID) handler funcs with all provider models.
func NewModelsHandlers() (http.HandlerFunc, http.HandlerFunc) {
	return modelsListHandler(allModels), modelsByIDHandler(allModels)
}

func modelsListHandler(models []types.Model) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(types.ModelsResponse{
			Object: "list",
			Data:   models,
		})
	}
}

func modelsByIDHandler(models []types.Model) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		for _, m := range models {
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
