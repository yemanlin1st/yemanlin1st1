.PHONY: test bench build vet fmt

test:
	go test ./...

bench:
	go test -bench=. -benchmem ./internal/balancer

build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/omegabalancer ./cmd/omegabalancer

vet:
	go vet ./...

fmt:
	gofmt -w cmd internal
