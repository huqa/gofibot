# makefile for gofibot

# Repository root directory:
root_dir := $(CURDIR)

BINARY := gofibot
VERSION ?= latest
PLATFORMS := windows linux darwin
os = $(word 1, $@)

.PHONY: $(PLATFORMS)
$(PLATFORMS):
	mkdir -p bin
	cd $(root_dir)/cmd/gofibot && GOOS=$(os) GOARCH=amd64 go build -o $(root_dir)/bin/$(BINARY)-$(VERSION)-$(os)-amd64
	$(if $(filter $(os), windows), cd $(root_dir)/bin/ && cmd //C ren "$(BINARY)-$(VERSION)-$(os)-amd64" "$(BINARY)-$(VERSION)-$(os)-amd64.exe")


.PHONY: release
	release: windows linux darwin

clean:
	cd $(root_dir)/bin/ && rm *
