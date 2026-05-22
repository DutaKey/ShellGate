package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dutakey/shellgate/internal/store"
	"github.com/dutakey/shellgate/internal/types"
)

type AdminHandler struct {
	keys *store.KeyStore
}

func NewAdminHandler(keys *store.KeyStore) *AdminHandler {
	return &AdminHandler{keys: keys}
}

func (h *AdminHandler) CreateKey() http.HandlerFunc {
	type request struct {
		Name string `json:"name"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
			writeJSONError(w, http.StatusBadRequest, "name is required", "invalid_request_error")
			return
		}

		key, err := h.keys.Create(req.Name)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to create key: "+err.Error(), "internal_error")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(key)
	}
}

func (h *AdminHandler) ListKeys() http.HandlerFunc {
	type response struct {
		Object string           `json:"object"`
		Data   []store.APIKey   `json:"data"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response{
			Object: "list",
			Data:   h.keys.List(),
		})
	}
}

func (h *AdminHandler) RevokeKey() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if !h.keys.Revoke(id) {
			writeJSONError(w, http.StatusNotFound, "key not found", "not_found")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(types.ErrorResponse{
			Error: types.APIError{
				Message: "key revoked",
				Type:    "success",
			},
		})
	}
}
