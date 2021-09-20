GO_LDFLAGS := -X `go list ./version`.Revision=`git rev-parse --short HEAD 2>/dev/null`
GO_GCFLAGS :=

default: build

build:
	go build -o payung -tags "$(BUILDTAGS)" -ldflags "$(GO_LDFLAGS)" $(GO_GCFLAGS)