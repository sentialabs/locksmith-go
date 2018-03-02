VERSION=0.0.1
GOPATH=$(CURDIR)
MODULE=locksmith
BIN=bin
DIST=dist

_TARGETS=\
	darwin-386 \
	darwin-amd64 \
	linux-386 \
	linux-amd64 \
	linux-arm \
	linux-arm64 \
	windows-386 \
	windows-amd64
TARGET_BINS=$(patsubst %,$(BIN)/$(MODULE)-%,$(_TARGETS))
TARGET_ZIPS=$(patsubst %,$(DIST)/$(MODULE)-%-$(VERSION).zip,$(_TARGETS))

default: $(TARGET_ZIPS)

clean:
	rm -rf $(BIN) $(DIST)

$(DIST)/%-$(VERSION).zip: $(BIN)/%
	@mkdir -p $(@D)
	zip $@ $(BIN)/$(*F)

$(BIN)/$(MODULE)-%: $(MODULE)/*.go
	name="$(@F)"; \
	module=$${name%%-*}; \
	os_arch=$${name#*-}; \
	os=$${os_arch%%-*}; \
	arch=$${os_arch##*-}; \
	GOOS=$$os GOARCH=$$arch go build -o $@ ./$$module

.PHONY: default
.SECONDARY: $(TARGET_BINS)
