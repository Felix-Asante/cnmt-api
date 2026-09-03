ifneq (,$(wildcard .env))
  include .env
  export
endif

migrate-up:
	goose -dir db/migrations postgres "$(DATABASE_URL)" up

# Same as Railway pre-deploy: runs migrations via the API binary (embedded SQL).
migrate-up-app:
	go run ./cmd/api migrate up

migrate-status:
	go run ./cmd/api migrate status

migrate-create:
	goose -dir db/migrations create $(name) sql

migrate-down:
	goose -dir db/migrations postgres "$(DATABASE_URL)" down

river-migrate-up:
	river migrate-up --database-url "$(DATABASE_URL)"

river-migrate-down:
	river migrate-down --line main --database-url "$(DATABASE_URL)" --max-steps 10