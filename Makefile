build:
	@go build ./...

.PHONY: update-index
update-index:
	@cd index && go run .
