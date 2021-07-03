BUILD=$(shell git rev-parse --short HEAD)
VERSION=$(shell git tag --points-at=$(BUILD); [ -z $$(git tag --points-at=$(BUILD)) ] && git branch --show-current | sed 's/\//_/g' )
MODULE=locksmith
BIN=bin
DIST=dist
GOPATH=$(CURDIR)
GOBIN=$(GOPATH)/$(BIN)

_TARGETS=\
	darwin-amd64 \
	linux-amd64 \
	linux-arm \
	linux-arm64 \
	windows-amd64

TARGET_BINS=$(patsubst %,$(BIN)/$(MODULE)-%,$(_TARGETS))
TARGET_ZIPS=$(patsubst %,$(DIST)/$(MODULE)-%-$(VERSION).zip,$(_TARGETS))

default: clean depend info build

clean:
	go clean
	rm -rf $(BIN) $(DIST)

depend:
	go get ./...

info: 
	@echo Version: $(VERSION)
	@echo Build: $(BUILD)

build: $(TARGET_ZIPS)

install:
	go install ./...

$(DIST)/%-$(VERSION).zip: $(BIN)/%
	@mkdir -p $(@D)
	zip $@ $(BIN)/$(*F)

$(BIN)/$(MODULE)-%: $(MODULE)/*.go
	name="$(@F)"; \
	module=$${name%%-*}; \
	os_arch=$${name#*-}; \
	os=$${os_arch%%-*}; \
	arch=$${os_arch##*-}; \
	GOOS=$$os GOARCH=$$arch \
	go build \
		-ldflags "-X=main.Version=$(VERSION) -X=main.Build=$(BUILD)" \
		-o $@ ./$$module

.PHONY: default
.SECONDARY: $(TARGET_BINS)
