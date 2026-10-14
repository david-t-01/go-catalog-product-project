package categories

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/david-t-01/go-catalog-product-project/app/api"
	"github.com/david-t-01/go-catalog-product-project/models"
)

const maxBodyBytes = 1 << 20

type Categories interface {
	List(ctx context.Context) ([]Category, error)
	Create(ctx context.Context, in CreateInput) (Category, error)
}

type Handler struct {
	svc Categories
	log *slog.Logger
}

func NewHandler(svc Categories, log *slog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

type createRequest struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.List(r.Context())
	if err != nil {
		h.writeError(w, r, "listing categories", err)
		return
	}
	api.OKResponse(w, ListResponse{Categories: list})
}

func (h *Handler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes)).Decode(&req); err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	c, err := h.svc.Create(r.Context(), CreateInput(req))
	if err != nil {
		h.writeError(w, r, "creating category", err)
		return
	}
	api.JSONResponse(w, http.StatusCreated, c)
}

// Internal error details are logged here and never reach the response body.
func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, op string, err error) {
	var ve *models.ValidationError
	switch {
	case errors.As(err, &ve):
		api.ErrorResponse(w, http.StatusBadRequest, ve.Error())
	case errors.Is(err, models.ErrConflict):
		api.ErrorResponse(w, http.StatusConflict, "category code already exists")
	default:
		h.log.ErrorContext(r.Context(), op, "err", err)
		api.ErrorResponse(w, http.StatusInternalServerError, "internal error")
	}
}
