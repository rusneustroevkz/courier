include cmd/backend/.env
export

sqlc-gen:
	sqlc generate

goose-up:
	goose up

goose-down:
	goose down

goose-create:
	goose -dir migrations create $(name) sql

docs-gen:
	swag init -g ./cmd/client/main.go --outputTypes go,json