export PROJECT_BRANCH  ?= $(shell git rev-parse --abbrev-ref HEAD)
export PROJECT_SHA     ?= $(shell git rev-parse HEAD)
export PROJECT_VERSION ?= v0.0.0-dev
export PROJECT_RELEASE ?= dev

export DEV_DOCKER_IMAGE_REPO_SIMULATOR 			?= gcr.io/aporetodev/simulator
export DEV_DOCKER_IMAGE_TAG       				?= $(PROJECT_VERSION)

MAKEFLAGS       += --warn-undefined-variables
SHELL           := /bin/bash -o pipefail

ARTIFACTS := docker/artifacts
SIMUTILS := $(UTILDIRS) docker/simulator/utils
CHARTDIRS := $(ARTIFACTS)/charts docker/simulator/charts

export GO111MODULE = on
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0

.PHONY: PHONY
PHONY = PHONY

build: $(PHONY) clean
	go mod tidy
	cd utils/plan-gen && go build
	cd utils/simulator && go build -o policies
	docker run --rm -v $(shell pwd)/utils/simulator/charts:/charts \
	  -v $(shell pwd)/docker:/docs \
	  alpine/helm package /charts/enforcer-sim -d /docs

simulator: build simulator-pkg
	docker build  -t ${DEV_DOCKER_IMAGE_REPO_SIMULATOR}:${DEV_DOCKER_IMAGE_TAG} docker/simulator

simulator-pkg:
	for DIR in $(SIMUTILS); do \
		mkdir -p $$DIR ; \
		cp utils/simulator/simulator.sh $$DIR/simulator.sh ; \
		cp utils/simulator/policies $$DIR/policies; \
		cp utils/plan-gen/plan-gen $$DIR/plan-gen; \
		chmod -R +x $$DIR ; \
	done
	for DIR in $(CHARTDIRS); do \
		mkdir -p $$DIR; \
		cp docker/enforcer-sim-*.tgz $$DIR; \
	done

clean:
	rm -rf docker/plan-gen/utils
	rm -rf docker/simulator/utils
	rm -rf $(ARTIFACTS)
	rm -f docker/enforcer-sim-*.tgz
	rm -rf $(CHARTDIRS)
