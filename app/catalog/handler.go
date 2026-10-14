package catalog

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/shopspring/decimal"

	"github.com/david-t-01/go-catalog-product-project/app/api"
	"github.com/david-t-01/go-catalog-product-project/models"
)

type Catalog interface {
	List(ctx context.Context, q ListQuery) (ListResult, error)
	Details(ctx context.Context, code string) (ProductDetails, error)
}

type Handler struct {
	svc Catalog
	log *slog.Logger
}

func NewHandler(svc Catalog, log *slog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	q, err := parseListQuery(r.URL.Query())
	if err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	res, err := h.svc.List(r.Context(), q)
	if err != nil {
		h.writeError(w, r, "listing products", err)
		return
	}
	api.OKResponse(w, res)
}

// Defaults apply only when a parameter is absent; the service clamps the values.
func parseListQuery(v url.Values) (ListQuery, error) {
	q := ListQuery{Limit: DefaultLimit, CategoryCode: v.Get("category")}

	if s := v.Get("offset"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil {
			return ListQuery{}, errors.New("invalid offset: must be an integer")
		}
		q.Offset = n
	}
	if s := v.Get("limit"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil {
			return ListQuery{}, errors.New("invalid limit: must be an integer")
		}
		q.Limit = n
	}
	if s := v.Get("price_less_than"); s != "" {
		d, err := decimal.NewFromString(s)
		if err != nil {
			return ListQuery{}, errors.New("invalid price_less_than: must be a decimal number")
		}
		q.PriceLessThan = &d
	}
	return q, nil
}

// Internal error details are logged here and never reach the response body.
func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, op string, err error) {
	var ve *models.ValidationError
	switch {
	case errors.As(err, &ve):
		api.ErrorResponse(w, http.StatusBadRequest, ve.Error())
	case errors.Is(err, models.ErrNotFound):
		api.ErrorResponse(w, http.StatusNotFound, "product not found")
	default:
		h.log.ErrorContext(r.Context(), op, "err", err)
		api.ErrorResponse(w, http.StatusInternalServerError, "internal error")
	}
}

func (h *Handler) HandleDetails(w http.ResponseWriter, r *http.Request) {
	res, err := h.svc.Details(r.Context(), r.PathValue("code"))
	if err != nil {
		h.writeError(w, r, "getting product", err)
		return
	}
	api.OKResponse(w, res)
}
