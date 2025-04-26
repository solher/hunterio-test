package dataextraction

import (
	"net/http"

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
