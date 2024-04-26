build:
	@go build ./...

.PHONY: update-index
update-index:
	@cd index && echo "" > sanctioned_addresses.json && go run .
