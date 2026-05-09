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
run-dev: run-templ run-tailwind run-esbuild-client

run-templ:
	docker stop templ && docker rm templ && docker run --name templ -v `pwd`:/app -w=/app ghcr.io/a-h/templ:latest generate

run-tailwind:
	npx @tailwindcss/cli -i ./frontend/styles/input.css -o ./frontend/styles/output.css -m

run-esbuild-client:
	./node_modules/.bin/esbuild --bundle ./frontend/client/index.tsx --outdir=public/client --minify --jsx=automatic --allow-overwrite --conditions=style