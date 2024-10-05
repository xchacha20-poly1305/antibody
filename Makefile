NAME = antibody
PARAMS = -v -trimpath -ldflags "-s -w -buildid="
MAIN = ./cmd/$(NAME)

.PHONY: build

build:
	go build $(PARAMS) $(MAIN)

fmt:
	@gofumpt -l -w .
	@gofmt -s -w .
	@gci write --custom-order -s standard -s "default" .

fmt_install:
	go install -v mvdan.cc/gofumpt@latest
	go install -v github.com/daixiang0/gci@latest