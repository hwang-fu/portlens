.PHONY: build clean test

build:
	go build -o portlens ./cmd/portlens

clean:
	rm -f portlens

test:
	go test ./...
