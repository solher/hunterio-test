package dataextraction

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/solher/toolbox/api"
)

// NewHTTPHandler returns a new HTTP handler for the service.
func NewHTTPHandler(service Service, json *api.JSON) http.Handler {
	h := &httpHandler{
		service: service,
		json:    json,
	}

	router := chi.NewRouter()
	router.Post("/", h.ExtractAndPersistFromURL)
	router.Post("/history", h.GetExtractedDataHistory)

	return router
}

type httpHandler struct {
	service Service
	json    *api.JSON
}

func (h *httpHandler) ExtractAndPersistFromURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	result, err := h.service.ExtractAndPersistFromURL(ctx, r.URL.Query().Get("url"))
	if err != nil {
		switch err {
		case ErrPageNotFound:
			h.json.RenderError(ctx, w, api.HTTPNotFound, err)
		default:
			h.json.RenderError(ctx, w, api.HTTPInternal, err)
		}
		return
	}

	h.json.Render(ctx, w, http.StatusOK, result)
}

func (h *httpHandler) GetExtractedDataHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		URL           string    `json:"url"`
		CreatedAtFrom time.Time `json:"created_at_from"`
		CreatedAtTo   time.Time `json:"created_at_to"`
		Limit         int       `json:"limit"`
		Offset        int       `json:"offset"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.json.RenderError(ctx, w, api.HTTPBodyDecoding, err)
		return
	}

	result, err := h.service.GetExtractedDataHistory(ctx, req.URL, req.CreatedAtFrom, req.CreatedAtTo, req.Limit, req.Offset)
	if err != nil {
		h.json.RenderError(ctx, w, api.HTTPInternal, err)
		return
	}

	h.json.Render(ctx, w, http.StatusOK, result)
}
