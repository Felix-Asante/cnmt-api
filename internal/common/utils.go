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
	now := time.Now()
	day := now.Day()
	month := now.Month()
	year := now.Year()

	entropy := ulid.Monotonic(rand.Reader, 0)
	id, err := ulid.New(ulid.Timestamp(now), entropy)
	if err != nil {
		return ""
	}

	reference := fmt.Sprintf("CNMT-%d%d%d-%s", year, month, day, id.String())
	return reference
}

func UuidPtrToPgtype(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

