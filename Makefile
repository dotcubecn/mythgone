PROJECT := Mythgone
BUILD_DIR := build
TIMESTAMP := TS$(shell date +%s)
TARGETS := x64 x86

GO111MODULE := on
CGO_ENABLED := 0

.DEFAULT_GOAL := all

.PHONY: all $(TARGETS) c i ct v b

all: ct $(TARGETS) v

$(TARGETS): ct
	@mkdir -p $(BUILD_DIR)/$(TIMESTAMP)
	@case $@ in \
		x64) ARCH=amd64 OUTPUT=x64 ;; \
		x86) ARCH=386 OUTPUT=x86 ;; \
	esac; \
	rsrc -manifest ./app.manifest -ico ./icon.ico -arch $$ARCH -o ./$$ARCH.syso; \
	GOOS=windows GOARCH=$$ARCH CGO_ENABLED=$(CGO_ENABLED) GO111MODULE=$(GO111MODULE) \
	go build -ldflags="-s -w -H windowsgui" -trimpath -o "$(BUILD_DIR)/$(TIMESTAMP)/$(PROJECT) $$OUTPUT.exe"; \
	rm -f ./$$ARCH.syso

b: ct $(TARGETS)

c:
	@rm -rf $(BUILD_DIR)

i:
	@go mod tidy
	@go install github.com/akavel/rsrc@latest

ct:
	@command -v go >/dev/null 2>&1 || (echo "go not found." && exit 1)
	@command -v rsrc >/dev/null 2>&1 || (echo "rsrc not found." && exit 1)
	@command -v git >/dev/null 2>&1 || (echo "git not found." && exit 1)

v:
	@mkdir -p $(BUILD_DIR)/$(TIMESTAMP)
	@./versioninfo.sh "$(BUILD_DIR)/$(TIMESTAMP)/versioninfo.rc"