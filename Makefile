BUILDENVVAR = CGO_ENABLED=0
REGISTRY = gcr.io/istio-state-metrics
TAG = $(shell git rev-parse HEAD)
# PKGS = $(shell go list ./... | grep -v /vendor/)
BuildDate = $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
Commit = $(shell git rev-parse --short HEAD)
PKG=github.com/douglas-reid/kube-state-metrics

IMAGE = $(REGISTRY)/istio-state-metrics

build:
	go build -o istio-state-metrics cmd/istio-state-metrics/main.go

TEMP_DIR := $(shell mktemp -d)

container:
	cp -r * $(TEMP_DIR)
	$(BUILDENVVAR) go build -o $(TEMP_DIR)/istio-state-metrics cmd/istio-state-metrics/main.go
	docker build -t $(IMAGE):$(TAG) $(TEMP_DIR)

push:
	docker push $(IMAGE):$(TAG)

.PHONY: build push