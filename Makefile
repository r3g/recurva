.PHONY: build test lint fmt vet clean install install-rdef install-all check

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
	rm -f recurva recurva-linux rdef

install:
	go install ./cmd/recurva

install-rdef:
	go install ./cmd/rdef

install-all: install install-rdef

check: fmt vet lint test
	@echo "All checks passed."
