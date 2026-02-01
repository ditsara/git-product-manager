.PHONY: build test install clean test-local

build:
	@./scripts/build.sh

test:
	@go test ./...

test-local: build
	@./scripts/test-local.sh

install:
	@go install ./cmd/pm

clean:
	@rm -rf bin/ sandbox/ dist/
	@echo "✓ Cleaned build artifacts"

all: clean build test
