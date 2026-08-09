package common

import (
	"fmt"
	"time"

	"crypto/rand"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/oklog/ulid/v2"
)


func GenerateReference() string {
	now := time.Now().UTC()
	entropy := ulid.Monotonic(rand.Reader, 0)
	id, err := ulid.New(ulid.Timestamp(now), entropy)
	if err != nil {
		return fmt.Sprintf("CNMT-%s", uuid.New().String())
	}
	return fmt.Sprintf("CNMT-%s", id.String())
}

func UuidPtrToPgtype(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

func UuidToPgtype(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func PgUuidToUuid(id pgtype.UUID) uuid.UUID {
	if !id.Valid {
		return uuid.Nil
	}
	return uuid.UUID(id.Bytes)
}