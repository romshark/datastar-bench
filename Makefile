.PHONY: dev gogen checkfmt checkgen

gogen:
	go generate ./...

dev:
	go run github.com/romshark/templier@latest

checkfmt:
	@unformatted=$$(go run mvdan.cc/gofumpt@latest -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "Files not gofumpt formatted:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

checkgen:
	@go generate ./...
	@if ! git diff --quiet --exit-code; then \
		echo "Generated files are not up to date. Run 'go generate ./...'"; \
		git --no-pager diff; \
		exit 1; \
	fi

lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest run ./...
