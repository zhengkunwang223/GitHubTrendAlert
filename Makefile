GOCMD=go
GOBUILD=$(GOCMD) build
GOARCH=$(shell go env GOARCH)
GOOS=$(shell go env GOOS )
BASE_PAH := $(shell pwd)
BUILD_PATH = $(BASE_PAH)/build
APPNAME=GithubTrendAlert

build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOBUILD)  -o $(BUILD_PATH)/${APPNAME} main.go