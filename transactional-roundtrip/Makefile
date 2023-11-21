APP ?= daemon
VERSION ?= $(shell git describe --tags)
COMMIT=$(shell git rev-parse HEAD)
ROOT=$(shell git rev-parse --show-toplevel)/transactional-roundtrip
DOCKERFILE ?= Dockerfile

# TAG is a sanitized version of the git tag, suitable for use as a docker tag
TAG ?= $(shell echo $${VERSION:="0.0.0-${COMMIT}"}|sed -E '1,$$s/[^a-zA-Z0-9_\.\-]/_/g'|sed -E '1,$$s/^[^a-zA-Z0-9_]/_/'|cut -c 1-128)

include deployment/deploy.mk

cli-doc:
	cd $(ROOT)/hack && go run gen_cli_doc.go


.PHONY: all-images
all-images: image injector-image

# This works with any docker
image: export BUILDER=DOCKER_BUILDKIT=1 docker build
image: export DOCKERFILE=Dockerfile
image: export DOCKER_TAG=$(TAG)
image:
	make build-image

# Same as image, but with an alpine build (the image is slightly smaller than with the distroless base image)
image-alpine: export BUILDER=DOCKER_BUILDKIT=1 docker build
image-alpine: export DOCKERFILE=Dockerfile.alpine
image-alpine: export DOCKER_TAG=$(TAG)-alpine
image-alpine:
	make build-image

# This requires the docker-buildx-plugin to be installed
imagex: export BUILDER=docker buildx build
imagex: export DOCKERFILE=Dockerfile
imagex: export DOCKER_TAG=$(TAG)
imagex:
	make build-image

injector-image: export BUILDER=DOCKER_BUILDKIT=1 docker build
injector-image: export DOCKERFILE=Dockerfile.injector
injector-image: export DOCKER_TAG=$(TAG)
injector-image: export APP=injector
injector-image:
	make build-image

build-image:
	@echo "building image for version ${VERSION}"
	cd $(ROOT)
	$(BUILDER) \
	--ssh default --progress auto --network host \
	--pull \
	--build-arg version=${DOCKER_TAG} \
	--build-arg commit=${COMMIT} \
	--sbom=true \
	--provenance=true \
	-t fredbi/${APP}:${DOCKER_TAG} \
	-f $(DOCKERFILE) \
	.
	
#--output type=local,dest=out
