.PHONY: build test install clean sandbox

build:
	@./scripts/build.sh

test:
	@go test ./...

sandbox: build
	@./scripts/sandbox.sh

install:
	@go install ./cmd/pm

clean:
	@rm -rf bin/ sandbox/ dist/
	@echo "✓ Cleaned build artifacts"

all: clean build test
