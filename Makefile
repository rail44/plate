.PHONY: test
test:
	go test ./... -count=1

.PHONY: test-verbose
test-verbose:
	go test ./... -count=1 -v

.PHONY: test-short
test-short:
	go test ./... -short

.PHONY: generate
generate:
	cd examples && go generate ./...

.PHONY: clean
clean:
	rm -rf examples/generated/