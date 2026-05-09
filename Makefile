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

.PHONY: run-dev
run-dev: run-templ run-esbuild-admin run-esbuild-client

run-templ:
	docker stop templ && docker rm templ && docker run --name templ -v `pwd`:/app -w=/app ghcr.io/a-h/templ:latest generate

run-esbuild-admin:
	./node_modules/.bin/esbuild --bundle ./static/admin/index.tsx --outdir=static/admin --minify --jsx=automatic --allow-overwrite

run-esbuild-client:
	./node_modules/.bin/esbuild --bundle ./static/admin/index.tsx --outdir=static/admin --minify --jsx=automatic --allow-overwrite