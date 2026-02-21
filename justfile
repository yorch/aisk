binary_name := "aisk"
build_dir := "bin"
version := `grep 'AppVersion' internal/config/config.go | head -1 | cut -d'"' -f2`
ldflags := "-s -w"

# Build binary to bin/aisk
build:
    mkdir -p {{build_dir}}
    go build -ldflags="{{ldflags}}" -o {{build_dir}}/{{binary_name}} ./cmd/aisk

# Build and copy to /usr/local/bin
install: build
    cp {{build_dir}}/{{binary_name}} /usr/local/bin/{{binary_name}}

# Run all tests with race detector
test:
    go test ./... -count=1 -race

# Run golangci-lint
lint:
    golangci-lint run ./...

# Remove build artifacts
clean:
    rm -rf {{build_dir}}

# GoReleaser snapshot build
snapshot:
    goreleaser release --snapshot --clean

# Run gofmt
fmt:
    gofmt -w .

# Run go vet
vet:
    go vet ./...

# Format, vet, and test
check: fmt vet test
