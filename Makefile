.PHONY: docker-build
IMG ?= ghcr.io/k8sgpt-ai/k8sgpt:latest

deploy:
ifndef SECRET
	$(error SECRET environment variable is not set)
endif
	kubectl create ns k8sgpt || true
	kubectl create secret generic ai-backend-secret --from-literal=secret-key=$(SECRET) --namespace=k8sgpt || true
	kubectl apply -f container/manifests
undeploy:
	kubectl delete secret ai-backend-secret --namespace=k8sgpt
	kubectl delete -f container/manifests
	kubectl delete ns k8sgpt
docker-build:
	docker buildx build --build-arg=VERSION="$$(git describe --tags --abbrev=0)" --build-arg=COMMIT="$$(git rev-parse --short HEAD)" --build-arg DATE="$$(date +%FT%TZ)" --platform="linux/amd64,linux/arm64" -t ${IMG} -f container/Dockerfile . --push