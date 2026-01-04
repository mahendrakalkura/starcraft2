.PHONY: build lint refresh reset sqlc

build:
	goimports -w .
	go mod tidy || (go get . && go mod tidy)
	go build .

lint:
	golangci-lint run ./...

refresh:
	./main --action=refresh

reset:
	psql postgres://postgres@0.0.0.0:5432/starcraft2 < sqlc/schema.sql

sqlc:
	sqlc generate
