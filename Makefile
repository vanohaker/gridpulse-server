tidy:
	go mod tidy

watch:
	air

generate:
	go generate cmd/main.go

migrate-up: