-include .env
export
version=
name=

docker-run-app:
	@docker compose up --build alohssh

create-migra:
	@goose -dir=$(MIGRATIONS_PATH) create $(name) sql

docker-migrate-up:
	@docker compose run --rm migrate up

docker-migrate-down:
	@docker compose run --rm migrate down