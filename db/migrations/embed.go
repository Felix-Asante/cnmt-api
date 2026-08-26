package migrations

import "embed"

// FS contains goose SQL migration files shipped inside the API binary.
//
//go:embed *.sql
var FS embed.FS
