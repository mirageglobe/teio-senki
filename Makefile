.DEFAULT_GOAL := help

.PHONY: help install data test tui build reload \
        export-mac export-linux export-windows export-android

help:                ## show this help menu
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*##"}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

install:             ## install dependencies (go mod tidy + yq via brew)
	brew install yq
	go mod tidy

data:                ## convert yaml archives to json (assets/data)
	@mkdir -p assets/data
	yq -o=json data/officers.yaml > assets/data/officers.json
	yq -o=json data/cities.yaml   > assets/data/cities.json

test:                ## run all Go tests headlessly
	go test ./...

tui:                 ## run the game in TUI mode
	go run ./cmd/teio

build:               ## build the binary to ./bin/teio
	@mkdir -p bin
	go build -o bin/teio ./cmd/teio

reload:              ## touch lua scripts to trigger hot-reload in a running TUI session
	find lua/ -name '*.lua' -exec touch {} +

export-mac:          ## cross-compile for macOS (arm64 + amd64 universal)
	GOOS=darwin  GOARCH=arm64  go build -o bin/teio-mac-arm64  ./cmd/teio
	GOOS=darwin  GOARCH=amd64  go build -o bin/teio-mac-amd64  ./cmd/teio

export-linux:        ## cross-compile for Linux (amd64)
	GOOS=linux   GOARCH=amd64  go build -o bin/teio-linux      ./cmd/teio

export-windows:      ## cross-compile for Windows (amd64)
	GOOS=windows GOARCH=amd64  go build -o bin/teio-windows.exe ./cmd/teio

export-android:      ## build Android APK via gomobile (requires gomobile setup)
	gomobile build -target android ./cmd/teio
