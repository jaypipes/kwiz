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

build-linux:
	GOARCH=amd64 GOOS=linux go build -o bin/${BINARY_NAME}-linux cmd/kwiz/main.go

build-darwin:
	GOARCH=amd64 GOOS=darwin go build -o bin/${BINARY_NAME}-darwin cmd/kwiz/main.go

build-windows:
	GOARCH=amd64 GOOS=windows go build -o bin/${BINARY_NAME}-windows cmd/kwiz/main.go

build-all: build-linux build-darwin build-windows

ifeq ($(detected_OS),Linux)
build: build-linux
endif
ifeq ($(detected_OS),Darwin)
build: build-darwin
endif
ifeq ($(detected_OS),Windows)
build: build-windows
endif

# If the first argument contains "run"...
ifeq (run,$(findstring run, $(firstword $(MAKECMDGOALS))))
  # use the rest as arguments for "run"
  RUN_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # ...and turn them into do-nothing targets
  $(eval $(RUN_ARGS):;@:)
endif

run-linux:
	./bin/${BINARY_NAME}-linux $(RUN_ARGS)

run-darwin:
	./bin/${BINARY_NAME}-darwin $(RUN_ARGS)

run-windows:
	./bin/${BINARY_NAME}-windows $(RUN_ARGS)

ifeq ($(detected_OS),Linux)
run: build run-linux
endif
ifeq ($(detected_OS),Darwin)
run: build run-darwin
endif
ifeq ($(detected_OS),Windows)
run: build run-windows
endif

build-image:
	docker build -t ${BINARY_NAME}:${VERSION} -f Dockerfile .

run-image:
	docker run --network host -it -v ${HOME}/.kube/config:/kconfig -e KUBECONFIG=/kconfig ${BINARY_NAME}:${VERSION}

clean:
	go clean
	rm bin/${BINARY_NAME}-darwin
	rm bin/${BINARY_NAME}-linux
	rm bin/${BINARY_NAME}-windows

test:
	go test ./...

test_coverage:
	go test ./... -coverprofile=coverage.out

dep:
	go mod download

vet:
	go vet

lint:
	golangci-lint run --enable-all
