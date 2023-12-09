BINARY_NAME=kwiz
VERSION ?= $(shell git describe --tags --always --dirty)

ifeq '$(findstring ;,$(PATH))' ';'
    detected_OS := Windows
else
    detected_OS := $(shell uname 2>/dev/null || echo Unknown)
    detected_OS := $(patsubst CYGWIN%,Cygwin,$(detected_OS))
    detected_OS := $(patsubst MSYS%,MSYS,$(detected_OS))
    detected_OS := $(patsubst MINGW%,MSYS,$(detected_OS))
endif

build:
	@GOARCH=amd64 GOOS=linux go build -o bin/${BINARY_NAME}-linux cmd/kwiz/main.go
	@GOARCH=amd64 GOOS=darwin go build -o bin/${BINARY_NAME}-darwin cmd/kwiz/main.go
	@GOARCH=amd64 GOOS=windows go build -o bin/${BINARY_NAME}-windows cmd/kwiz/main.go

run: build
ifeq ($(detected_OS),Linux)
	@./bin/${BINARY_NAME}-linux
endif
ifeq ($(detected_OS),Darwin)
	@./bin/${BINARY_NAME}-darwin
endif
ifeq ($(detected_OS),Windows)
	@./bin/${BINARY_NAME}-windows
endif

build-image:
	docker build -t ${BINARY_NAME}:${VERSION} -f Dockerfile .

clean:
	@go clean
	@rm bin/${BINARY_NAME}-darwin
	@rm bin/${BINARY_NAME}-linux
	@rm bin/${BINARY_NAME}-windows

test:
	@go test ./...

test_coverage:
	@go test ./... -coverprofile=coverage.out

dep:
	@go mod download

vet:
	@go vet

lint:
	@golangci-lint run --enable-all
