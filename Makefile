.PHONY: docker-build
IMG ?= ghcr.io/k8sgpt-ai/k8sgpt:latest
docker-build:
	docker buildx build --build-arg=VERSION="$$(git describe --tags --abbrev=0)" --build-arg=COMMIT="$$(git rev-parse --short HEAD)" --build-arg DATE="$$(date +%FT%TZ)" --platform="linux/amd64,linux/arm64" -t ${IMG} -f container/Dockerfile . --push