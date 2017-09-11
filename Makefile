.PHONY: cli test

cli:
	go install ./cmd/...

test:
	go test ./...


