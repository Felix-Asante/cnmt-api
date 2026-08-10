ifneq (,$(wildcard .env))
  include .env
  export
endif

migrate-up:
	goose -dir db/migrations postgres "$(DATABASE_URL)" up

migrate-create:
	goose -dir db/migrations create $(name) sql

migrate-down:
	goose -dir db/migrations postgres "$(DATABASE_URL)" down