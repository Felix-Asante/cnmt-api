package transfers

import (
	"fmt"
	"net/http"
	"strconv"

	"cnmt/internal/common/httpx"
	"cnmt/internal/infra/db"

	"github.com/google/uuid"
)

func parseGetAllTransfersQuery(r *http.Request) (getAllTransfersRequest, error) {
	q := r.URL.Query()
	req := getAllTransfersRequest{}

	if v := q.Get("sender_phone"); v != "" {
		req.SenderPhone = &v
	}
	if v := q.Get("recipient_phone"); v != "" {
		req.RecipientPhone = &v
	}
	if v := q.Get("reference"); v != "" {
		req.Reference = &v
	}
	if v := q.Get("status"); v != "" {
		status := db.TransferStatus(v)
		req.Status = &status
	}
	if v := q.Get("route_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return req, fmt.Errorf("%w: invalid route_id", httpx.BadRequestError)
		}
		req.RouteID = &id
	}
	if v := q.Get("page"); v != "" {
		page, err := strconv.Atoi(v)
		if err != nil {
			return req, fmt.Errorf("%w: invalid page", httpx.BadRequestError)
		}
		req.Page = &page
	}
	if v := q.Get("limit"); v != "" {
		limit, err := strconv.Atoi(v)
		if err != nil {
			return req, fmt.Errorf("%w: invalid limit", httpx.BadRequestError)
		}
		req.Limit = &limit
	}

	return httpx.Validate(req)
}
