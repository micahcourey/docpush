# docpush Makefile

VERSION ?= dev
LDFLAGS := -s -w -X main.version=$(VERSION)
DIST    := dist

# --- Default -----------------------------------------------------------

.PHONY: all
all: test build

# --- Build (current platform) ------------------------------------------

.PHONY: build
build:
	go build -buildvcs=false -ldflags="$(LDFLAGS)" -o $(DIST)/docpush ./cmd/docpush

# --- Cross-compile ------------------------------------------------------

.PHONY: dist
dist:
	GOOS=linux   GOARCH=amd64 go build -buildvcs=false -ldflags="$(LDFLAGS)" -o $(DIST)/docpush-linux-amd64   ./cmd/docpush
	GOOS=darwin  GOARCH=amd64 go build -buildvcs=false -ldflags="$(LDFLAGS)" -o $(DIST)/docpush-darwin-amd64  ./cmd/docpush
	GOOS=darwin  GOARCH=arm64 go build -buildvcs=false -ldflags="$(LDFLAGS)" -o $(DIST)/docpush-darwin-arm64  ./cmd/docpush

# --- Zip (for release) --------------------------------------------------

.PHONY: zip
zip: dist
	@mkdir -p $(DIST)/zip
	@cp $(DIST)/docpush-linux-amd64 $(DIST)/zip/docpush && \
	  cd $(DIST)/zip && zip -q docpush-linux-amd64.zip docpush && rm docpush
	@cp $(DIST)/docpush-darwin-amd64 $(DIST)/zip/docpush && \
	  cd $(DIST)/zip && zip -q docpush-darwin-amd64.zip docpush && rm docpush
	@cp $(DIST)/docpush-darwin-arm64 $(DIST)/zip/docpush && \
	  cd $(DIST)/zip && zip -q docpush-darwin-arm64.zip docpush && rm docpush

# --- Test & Lint --------------------------------------------------------

.PHONY: test lint vet
test:
	go test ./...

lint: vet
vet:
	go vet ./...

# --- Clean --------------------------------------------------------------

.PHONY: clean
clean:
	rm -rf $(DIST)
