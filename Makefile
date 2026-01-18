export CGO_ENABLED:=0

DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
VERSION=$(shell git describe --tags --match=v* --always --dirty)
LD_FLAGS="-w -X github.com/homeric-io/tinm/tinm/version.Version=$(VERSION)"

REPO=github.com/homeric-io/tinm
LOCAL_REPO=homeric-io/tinm
IMAGE_REPO=docker.io/aalaesar/tinm

.PHONY: all
all: build test vet fmt

.PHONY: build
build:
	@go build -o bin/tinm -ldflags $(LD_FLAGS) $(REPO)/cmd/tinm

.PHONY: test
test:
	@go test ./... -cover

.PHONY: vet
vet:
	@go vet -all ./...

.PHONY: fmt
fmt:
	@test -z $$(go fmt ./...)

.PHONY: lint
lint:
	@golangci-lint run ./...

.PHONY: image
image: \
	image-amd64 \
	image-arm64

image-%:
	buildah bud -f Dockerfile \
	-t $(LOCAL_REPO):$(VERSION)-$* \
	--arch $* --override-arch $* \
	--format=docker .

push: \
	push-amd64
	push-arm64

push-%:
	buildah tag $(LOCAL_REPO):$(VERSION)-$* $(IMAGE_REPO):$(VERSION)-$*
	buildah push --format v2s2 $(IMAGE_REPO):$(VERSION)-$*

manifest:
	buildah manifest create $(IMAGE_REPO):$(VERSION)
	buildah manifest add $(IMAGE_REPO):$(VERSION) docker://$(IMAGE_REPO):$(VERSION)-amd64
	buildah manifest add --variant v8 $(IMAGE_REPO):$(VERSION) docker://$(IMAGE_REPO):$(VERSION)-arm64
	buildah manifest inspect $(IMAGE_REPO):$(VERSION)
	buildah manifest push -f v2s2 $(IMAGE_REPO):$(VERSION) docker://$(IMAGE_REPO):$(VERSION)

protoc/%:
	podman run --security-opt label=disable \
		-u root \
		--mount type=bind,src=$(DIR),target=/mnt/code \
		quay.io/dghubble/protoc:v3.10.1 \
		--go_out=plugins=grpc,paths=source_relative:. $*

codegen: \
	protoc/tinm/storage/storagepb/*.proto \
	protoc/tinm/server/serverpb/*.proto \
	protoc/tinm/rpc/rpcpb/*.proto

clean:
	@rm -rf bin

clean-release:
	@rm -rf _output

release: \
	clean \
	clean-release \
	_output/tinm-linux-amd64.tar.gz \
	_output/tinm-linux-arm.tar.gz \
	_output/tinm-linux-arm64.tar.gz \
	_output/tinm-darwin-amd64.tar.gz \
	_output/tinm-darwin-arm64.tar.gz

bin/linux-amd64/tinm: GOARGS = GOOS=linux GOARCH=amd64
bin/linux-arm/tinm: GOARGS = GOOS=linux GOARCH=arm GOARM=6
bin/linux-arm64/tinm: GOARGS = GOOS=linux GOARCH=arm64
bin/darwin-amd64/tinm: GOARGS = GOOS=darwin GOARCH=amd64
bin/darwin-arm64/tinm: GOARGS = GOOS=darwin GOARCH=arm64
bin/linux-ppc64le/tinm: GOARGS = GOOS=linux GOARCH=ppc64le

bin/%/tinm:
	$(GOARGS) go build -o $@ -ldflags $(LD_FLAGS) -a $(REPO)/cmd/tinm

_output/tinm-%.tar.gz: NAME=tinm-$(VERSION)-$*
_output/tinm-%.tar.gz: DEST=_output/$(NAME)
_output/tinm-%.tar.gz: bin/%/tinm
	mkdir -p $(DEST)
	cp bin/$*/tinm $(DEST)
	./scripts/dev/release-files $(DEST)
	tar zcvf $(DEST).tar.gz -C _output $(NAME)

.PHONY: all build clean test release
.SECONDARY: _output/tinm-linux-amd64 _output/tinm-darwin-amd64

release-sign:
	gpg2 --armor --detach-sign _output/tinm-$(VERSION)-linux-amd64.tar.gz
	gpg2 --armor --detach-sign _output/tinm-$(VERSION)-linux-arm.tar.gz
	gpg2 --armor --detach-sign _output/tinm-$(VERSION)-linux-arm64.tar.gz
	gpg2 --armor --detach-sign _output/tinm-$(VERSION)-darwin-amd64.tar.gz
	gpg2 --armor --detach-sign _output/tinm-$(VERSION)-darwin-arm64.tar.gz

release-verify: NAME=_output/tinm
release-verify:
	gpg2 --verify $(NAME)-$(VERSION)-linux-amd64.tar.gz.asc $(NAME)-$(VERSION)-linux-amd64.tar.gz
	gpg2 --verify $(NAME)-$(VERSION)-linux-arm.tar.gz.asc $(NAME)-$(VERSION)-linux-arm.tar.gz
	gpg2 --verify $(NAME)-$(VERSION)-linux-arm64.tar.gz.asc $(NAME)-$(VERSION)-linux-arm64.tar.gz
	gpg2 --verify $(NAME)-$(VERSION)-darwin-amd64.tar.gz.asc $(NAME)-$(VERSION)-darwin-amd64.tar.gz
	gpg2 --verify $(NAME)-$(VERSION)-darwin-arm64.tar.gz.asc $(NAME)-$(VERSION)-darwin-arm64.tar.gz
