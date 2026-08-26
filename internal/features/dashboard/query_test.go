package dashboard

import (
	"errors"
	"net/http"
	"net/url"
	"testing"
	"time"

	"cnmt/internal/common/httpx"
)

func TestParsePeriod_DefaultCurrentMonth(t *testing.T) {
	now := time.Date(2026, 8, 25, 15, 30, 0, 0, time.UTC)
	req := &http.Request{URL: &url.URL{RawQuery: ""}}

	period, err := parsePeriod(req, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantFrom := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	wantTo := time.Date(2026, 8, 26, 0, 0, 0, 0, time.UTC)
	if !period.From.Equal(wantFrom) || !period.To.Equal(wantTo) {
		t.Fatalf("got period %+v, want from=%s to=%s", period, wantFrom, wantTo)
	}
}

func TestParsePeriod_CustomRange(t *testing.T) {
	now := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)
	req := &http.Request{URL: &url.URL{RawQuery: "from=2026-08-01&to=2026-08-25"}}

	period, err := parsePeriod(req, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantFrom := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	wantTo := time.Date(2026, 8, 26, 0, 0, 0, 0, time.UTC)
	if !period.From.Equal(wantFrom) || !period.To.Equal(wantTo) {
		t.Fatalf("got period %+v, want from=%s to=%s", period, wantFrom, wantTo)
	}
}

func TestParsePeriod_InvalidCases(t *testing.T) {
	now := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name  string
		query string
	}{
		{name: "only from", query: "from=2026-08-01"},
		{name: "only to", query: "to=2026-08-25"},
		{name: "bad from", query: "from=08-01-2026&to=2026-08-25"},
		{name: "from after to", query: "from=2026-08-25&to=2026-08-01"},
		{name: "too large", query: "from=2024-01-01&to=2026-08-25"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := &http.Request{URL: &url.URL{RawQuery: tc.query}}
			_, err := parsePeriod(req, now)
			if !errors.Is(err, httpx.BadRequestError) {
				t.Fatalf("expected BadRequestError, got %v", err)
			}
		})
	}
}
