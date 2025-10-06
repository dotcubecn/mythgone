PROJECT := Mythgone
BUILD_DIR := build
TIMESTAMP := TS$(shell date +%s)
TARGETS := x64 x86

GO111MODULE := on
CGO_ENABLED := 1

.DEFAULT_GOAL := all

.PHONY: all $(TARGETS) c i ct v b r

all: ct $(TARGETS) v

r: ct $(TARGETS) v
	@cp "$(BUILD_DIR)/$(TIMESTAMP)/$(PROJECT) x64.exe" "$(BUILD_DIR)/$(TIMESTAMP)/$(PROJECT) x64 NoUPX.exe"
	@cp "$(BUILD_DIR)/$(TIMESTAMP)/$(PROJECT) x86.exe" "$(BUILD_DIR)/$(TIMESTAMP)/$(PROJECT) x86 NoUPX.exe"
	@upx --best --lzma "$(BUILD_DIR)/$(TIMESTAMP)/$(PROJECT) x64.exe"
	@upx --best --lzma "$(BUILD_DIR)/$(TIMESTAMP)/$(PROJECT) x86.exe"

$(TARGETS): ct
	@mkdir -p $(BUILD_DIR)/$(TIMESTAMP)
	@case $@ in \
		x64) ARCH=amd64 OUTPUT=x64 CC=x86_64-w64-mingw32-gcc ;; \
		x86) ARCH=386 OUTPUT=x86 CC=i686-w64-mingw32-gcc ;; \
	esac; \
	rsrc -manifest ./app.manifest -ico ./icon.ico -arch $$ARCH; \
	GOOS=windows GOARCH=$$ARCH CGO_ENABLED=$(CGO_ENABLED) CC=$$CC CXX=$${CC/gcc/g++} GO111MODULE=$(GO111MODULE) \
	go build $(if $(filter r,$(MAKECMDGOALS)),-a) -tags walk_use_cgo -ldflags="-s -w -H windowsgui -X main.appName=Mythgone -buildmode=exe -extldflags=-static" -trimpath -buildvcs=false -o "$(BUILD_DIR)/$(TIMESTAMP)/$(PROJECT) $$OUTPUT.exe"; \
	rm -f ./*.syso

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
	@command -v x86_64-w64-mingw32-gcc >/dev/null 2>&1 || (echo "x86_64-w64-mingw32-gcc not found." && exit 1)
	@command -v i686-w64-mingw32-gcc >/dev/null 2>&1 || (echo "i686-w64-mingw32-gcc not found." && exit 1)
	@command -v upx >/dev/null 2>&1 || (echo "upx not found." && exit 1)

v:
	@mkdir -p $(BUILD_DIR)/$(TIMESTAMP)
	@./versioninfo.sh "$(BUILD_DIR)/$(TIMESTAMP)/versioninfo.rc"