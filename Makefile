.PHONY: build test lint fmt vet clean install

build:
	go build ./...

test:
	go test ./... -count=1

lint:
	golangci-lint run ./...

fmt:
	gofumpt -w .

vet:
	go vet ./...

clean:
	rm -f recurva recurva-linux

install:
	go install ./cmd/recurva

check: fmt vet lint test
	@echo "All checks passed."
