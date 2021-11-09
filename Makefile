PKG:=github.com/kayrus/ec2auth
APP_NAME:=ec2auth
PWD:=$(shell pwd)
UID:=$(shell id -u)
VERSION:=$(shell git describe --tags --always --dirty="-dev")
LDFLAGS:=-X $(PKG)/pkg.Version=$(VERSION)

export CGO_ENABLED:=0

build: fmt linux darwin windows

linux:
	GOOS=linux go build -ldflags="$(LDFLAGS)" -o bin/$(APP_NAME) ./cmd

darwin:
	GOOS=darwin go build -ldflags="$(LDFLAGS)" -o bin/$(APP_NAME)_darwin ./cmd

windows:
	GOOS=windows go build -ldflags="$(LDFLAGS)" -o bin/$(APP_NAME).exe ./cmd

docker:
	docker run -ti --rm -e GOCACHE=/tmp -v $(PWD):/$(APP_NAME) -u $(UID):$(UID) --workdir /$(APP_NAME) golang:latest make

fmt:
	gofmt -s -w cmd pkg
