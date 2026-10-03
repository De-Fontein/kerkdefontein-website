.PHONY: js test
js:
	npx tsc

test: js
	go vet ./...
	go test ./...
