GIT_TAG?=$(shell git describe --candidates=50 --abbrev=0 --tags 2>/dev/null || echo "v0.0.1" )
GIT_COMMIT?=$(shell git rev-parse HEAD)
GIT_COMMIT_SHORT?=$(shell git rev-parse --short HEAD)
GO_MODULE?= $(shell go list -m)
DOCKERIMG?=quotaday
DOCKERTAG?=${GIT_TAG}

LDFLAGS:=-w -s
LDFLAGS+=-X "$(GO_MODULE)/cmd.version=$(GIT_TAG)"
LDFLAGS+=-X "$(GO_MODULE)/cmd.gitCommit=$(GIT_COMMIT)"

.PHONY: clean
clean:
	rm -rf quotaday

.PHONY: generate
generate:
	go generate ./...

.PHONY: quotaday
quotaday:
	CGO_ENABLED=0 go build -ldflags '$(LDFLAGS)' -o quotaday main.go

.PHONY: docker
docker:
	DOCKER_BUILDKIT=1 docker build \
			-f Dockerfile \
			--build-arg VERSION=${GIT_TAG} \
			--build-arg COMMIT=${GIT_COMMIT_SHORT} \
			-t ${DOCKERIMG}:${DOCKERTAG}
