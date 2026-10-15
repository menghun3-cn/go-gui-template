VERSION ?= $(shell git describe --tags --always --dirty 2>nul || echo dev)
GO ?= $(shell where go 2>nul || echo go)

.PHONY: build build-gui build-cli clean run test

build:
	go build -tags production -trimpath -ldflags "-s -w -X main.version=$(VERSION) -H windowsgui" -o dist/gogui.exe ./cmd/gogui

build-cli:
	go build -tags nogui -trimpath -ldflags "-s -w -X main.version=$(VERSION)" -o dist/gogui-cli.exe ./cmd/gogui

run:
	go run -tags production ./cmd/gogui

test:
	go test ./...

clean:
	rm -rf dist
