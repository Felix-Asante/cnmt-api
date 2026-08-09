package common

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

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

	
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w: %v", httpx.NotFoundError, err)
	}

	if errors.Is(err, context.Canceled) {
		return fmt.Errorf("%w: %v", errors.New("request canceled"), err)
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("%w: %v", errors.New("request timeout"), err)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return translatePgError(pgErr)
	}

	return fmt.Errorf("%w", httpx.InternalServerError)
}

func translatePgError(pgErr *pgconn.PgError) error {
	switch pgErr.Code {
	case pgUniqueViolation:
		return fmt.Errorf("%w: %s", errors.New("Record already exists"), constraintDetail(pgErr))
	case pgNotNullViolation, pgCheckViolation:
		return fmt.Errorf("%w: %s", errors.New("Required field is missing"), constraintDetail(pgErr))
	case pgForeignKeyViolation:
		return fmt.Errorf("%w: %s", errors.New("foreign key violation"), constraintDetail(pgErr))
	default:
		slog.Error("unhandled postgres error code",
			"code", pgErr.Code,
			"message", pgErr.Message,
			"constraint", pgErr.ConstraintName,
		)
		return fmt.Errorf("%w", httpx.InternalServerError)
	}
}

func constraintDetail(pgErr *pgconn.PgError) string {
	if pgErr.ConstraintName != "" {
		return pgErr.ConstraintName
	}
	return pgErr.Message
}