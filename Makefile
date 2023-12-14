GOARCH=$(shell go env GOARCH)
GOOS=$(shell go env GOOS )

build:
	GOOS=$(GOARCH) GOARCH=$(GOOS) go build  -o ./build/GithubTrendAlert main.go