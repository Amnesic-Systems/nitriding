prog = nitriding
src_dir = src
godeps = $(src_dir)/*.go $(src_dir)/go.mod $(src_dir)/go.sum
cover_out = cover.out
cover_html = cover.html

all: $(prog)

.PHONY: lint
lint: $(godeps)
	go vet -C $(src_dir) ./...
	govulncheck -C $(src_dir) ./...

.PHONY: test
test: $(godeps)
	go test -C $(src_dir) -race -cover ./...

.PHONY: coverage
coverage: $(cover_html)
	open $(src_dir)/$(cover_html)

$(cover_out): $(godeps)
	go test -C $(src_dir) -coverprofile=$(cover_out)

$(cover_html): $(cover_out)
	go tool -C $(src_dir) cover -html=$(cover_out) -o $(cover_html)

$(prog): $(godeps)
	CGO_ENABLED=0 go build \
		-C $(src_dir) \
		-trimpath \
		-ldflags="-s -w" \
		-buildvcs=false \
		-o $(prog)

.PHONY: clean
clean:
	rm -f $(src_dir)/$(prog)
	rm -f $(src_dir)/$(cover_out)
	rm -f $(src_dir)/$(cover_html)
