.PHONY: all build build-widget build-web test clean run

VERSION ?= 1.0.0
COMMIT ?= $(shell git rev-parse --short HEAD 2> /dev/null || echo "unknown")
LDFLAGS := -s -w -X main.Version=$(VERSION) -X main.Commit=$(COMMIT)

all: build

build-widget:
	@echo "Building standalone search widget..."
	cd widget && npm install && npm run build

build-web:
	@echo "Building admin studio web UI..."
	cd web && npm install && npm run build

build: build-widget build-web
	@echo "Building SearchCraft binary..."
	go build -ldflags="$(LDFLAGS)" -o bin/searchcraft ./cmd/searchcraft

test:
	go test -v -race ./internal/...

clean:
	rm -rf bin/ dist/ widget/dist/ web/dist/ *.db*

run: build
	./bin/searchcraft serve --port 7700 --db ./searchcraft.db
