package dashboard

import (
	"fmt"
	"net/http"
	"time"

	"cnmt/internal/common/httpx"
)

func parsePeriod(r *http.Request, now time.Time) (Period, error) {
	now = now.UTC()
	q := r.URL.Query()
	fromRaw := q.Get("from")
	toRaw := q.Get("to")

	if fromRaw == "" && toRaw == "" {
		from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		toExclusive := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, 1)
		return Period{From: from, To: toExclusive}, nil
	}
	if fromRaw == "" || toRaw == "" {
		return Period{}, fmt.Errorf("%w: both from and to are required", httpx.BadRequestError)
	}

	fromDay, err := time.ParseInLocation("2006-01-02", fromRaw, time.UTC)
	if err != nil {
		return Period{}, fmt.Errorf("%w: invalid from date, use YYYY-MM-DD", httpx.BadRequestError)
	}
	toDay, err := time.ParseInLocation("2006-01-02", toRaw, time.UTC)
	if err != nil {
		return Period{}, fmt.Errorf("%w: invalid to date, use YYYY-MM-DD", httpx.BadRequestError)
	}
	if fromDay.After(toDay) {
		return Period{}, fmt.Errorf("%w: from must be on or before to", httpx.BadRequestError)
	}

	toExclusive := toDay.AddDate(0, 0, 1)
	if toExclusive.Sub(fromDay) > time.Duration(maxRangeDays)*24*time.Hour {
		return Period{}, fmt.Errorf("%w: date range cannot exceed %d days", httpx.BadRequestError, maxRangeDays)
	}

	return Period{From: fromDay, To: toExclusive}, nil
}
