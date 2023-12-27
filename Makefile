prog = nitriding
src_dir = internal
godeps = $(src_dir)/*.go *.go go.mod go.sum
cover_out = cover.out
cover_html = cover.html

all: $(prog)

.PHONY: lint
lint: $(godeps)
	go vet ./...
	govulncheck ./...

.PHONY: test
test: $(godeps)
	go test -race -cover ./...

.PHONY: coverage
coverage: $(cover_html)
	open $(src_dir)/$(cover_html)

$(cover_out): $(godeps)
	go test -C $(src_dir) -coverprofile=$(cover_out)

$(cover_html): $(cover_out)
	go tool -C $(src_dir) cover -html=$(cover_out) -o $(cover_html)

$(prog): $(godeps)
	CGO_ENABLED=0 go build \
		-trimpath \
		-ldflags="-s -w" \
		-buildvcs=false \
		-o $(prog)

.PHONY: clean
clean:
	rm -f $(prog)
	rm -f $(src_dir)/$(cover_out)
	rm -f $(src_dir)/$(cover_html)
