package common

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"cnmt/internal/common/httpx"
)

const (
	pgUniqueViolation     = "23505"
	pgNotNullViolation    = "23502"
	pgForeignKeyViolation = "23503"
	pgCheckViolation      = "23514"
	
)

func TranslateDBError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w", httpx.NotFoundError)
	}

	if errors.Is(err, context.Canceled) {
		return fmt.Errorf("%w", httpx.BadRequestError)
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("%w", httpx.GatewayTimeoutError)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return translatePgError(pgErr)
	}

	slog.Error("unhandled database error", "err", err)
	return fmt.Errorf("%w", httpx.InternalServerError)
}

func translatePgError(pgErr *pgconn.PgError) error {
	switch pgErr.Code {
	case pgUniqueViolation:
		slog.Warn("unique violation", "constraint", pgErr.ConstraintName)
		return fmt.Errorf("%w", httpx.ConflictError)
	case pgNotNullViolation, pgCheckViolation:
		slog.Warn("constraint violation", "code", pgErr.Code, "constraint", pgErr.ConstraintName)
		return fmt.Errorf("%w", httpx.BadRequestError)
	case pgForeignKeyViolation:
		slog.Warn("foreign key violation", "constraint", pgErr.ConstraintName)
		return fmt.Errorf("%w", httpx.BadRequestError)
	default:
		slog.Error("unhandled postgres error code",
			"code", pgErr.Code,
			"message", pgErr.Message,
			"constraint", pgErr.ConstraintName,
		)
		return fmt.Errorf("%w", httpx.InternalServerError)
	}
}
