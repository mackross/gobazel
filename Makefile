VERSION=0.1.6

all:
	@echo "make binary: build gobazel binary"

binary:
	go install -ldflags "-s -w -X main.version=${VERSION}" github.com/linuxerwang/gobazel
