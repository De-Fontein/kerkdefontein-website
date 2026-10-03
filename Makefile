.PHONY: js test
js:
	npx tsc

test: js
	go vet ./...
	go test ./...

VERSION := $(shell git describe --always --dirty 2>/dev/null || echo dev)
HOST ?= fontein

.PHONY: preview e2e build deploy
preview: js
	go run ./cmd/fontein-preview

e2e: js
	npx playwright test

build: js
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=$(VERSION)" -o dist/fontein-build ./cmd/fontein-build

deploy: build
	scp dist/fontein-build $(HOST):/tmp/fontein-build
	ssh $(HOST) 'sudo install -m 0755 /tmp/fontein-build /usr/local/bin/fontein-build && sudo systemctl start fontein-build.service && systemctl status --no-pager fontein-build.service | tail -n 5'
