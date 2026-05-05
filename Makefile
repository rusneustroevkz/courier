#include .env
#export

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