include cmd/backend/.env
export

.PHONY: sqlc-gen
sqlc-gen:
	sqlc generate

.PHONY: goose-up
goose-up:
	goose up

.PHONY: goose-down
goose-down:
	goose down

.PHONY: goose-create
goose-create:
	goose -dir migrations create $(name) sql

.PHONY: run-templ
run-templ:
	docker stop templ && docker rm templ && docker run --name templ -v `pwd`/public:/app -w=/app ghcr.io/a-h/templ:latest generate