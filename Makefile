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
	docker stop templ && docker rm templ && docker run --name templ -v `pwd`:/app -w=/app ghcr.io/a-h/templ:latest generate

.PHONY: run-esbuild
run-esbuild:
	./node_modules/.bin/esbuild --bundle static/app/index.js --outdir=static --minify --jsx=automatic