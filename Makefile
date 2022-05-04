# export PROJECT_BRANCH  ?= $(shell git rev-parse --abbrev-ref HEAD)
# export PROJECT_SHA     ?= $(shell git rev-parse HEAD)
export PROJECT_BRANCH  ?= testbranch
export PROJECT_SHA     ?= 41e8dd18a959770c6aa2e278cc7f36e41a109f69
export PROJECT_VERSION ?= v0.0.0-dev
export PROJECT_RELEASE ?= dev

# export DEV_DOCKER_IMAGE_REPO_BENCHMARK_SUITE    ?= gcr.io/aporetodev/benchmark-suite
export DEV_DOCKER_IMAGE_REPO_SIMULATOR 			?= gcr.io/aporetodev/simulator
export DEV_DOCKER_IMAGE_TAG       				?= $(PROJECT_VERSION)

MAKEFLAGS       += --warn-undefined-variables
SHELL           := /bin/bash -o pipefail

ARTIFACTS := docker/artifacts
# TESTDIRS := $(ARTIFACTS)/tests docker/benchmark-suite/tests
# UTILDIRS := $(ARTIFACTS)/utils docker/benchmark-suite/utils
SIMUTILS := $(UTILDIRS) docker/simulator/utils
CHARTDIRS := $(ARTIFACTS)/charts docker/simulator/charts

export GO111MODULE = on
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0

.PHONY: PHONY
PHONY = PHONY

# ci: build pkg
# 	@rm -rf artifacts/
# 	@mkdir -p artifacts/
# 	@echo "$(PROJECT_SHA)" > artifacts/src_sha
# 	@echo "$(PROJECT_VERSION)" > artifacts/src_semver
# 	@echo "$(PROJECT_BRANCH)" > artifacts/src_branch
# 	@if [[ -f Gopkg.toml ]] ; then cp Gopkg.toml artifacts/ ; fi
# 	@if [[ -f Gopkg.lock ]] ; then cp Gopkg.lock artifacts/ ; fi
# 	@if [[ -d build/ ]] ; then cp -r build/ artifacts/build/ ; fi
# 	@if [[ -d docker/ ]] ; then cp -r docker/ artifacts/docker/ ; fi
# 	@mkdir -p artifacts/repo/helm/
# 	@if [[ -d helm/repo/ ]] ; then cp -r helm/repo/* artifacts/repo/helm/ ; fi
# 	@if [[ -d helm/aggregated/ ]] ; then cp -r helm/aggregated/* artifacts/repo/helm/ ; fi
# 	@mkdir -p artifacts/repo/swarm/
# 	@if [[ -d swarm/repo/ ]] ; then cp -r swarm/repo/* artifacts/repo/swarm/ ; fi
# 	@if [[ -d swarm/aggregated/ ]] ; then cp -r swarm/aggregated/* artifacts/repo/swarm/ ; fi

build: $(PHONY) clean
	go mod tidy
# cd tests/control && go build
# cd tests/throughput && go build
# cd tests/pod-scaling && go build
# cd tests/latency && go build
# cd tests/windows && go build
# cd tests/enforcer && go build -o enforcerbench
	cd utils/plan-gen && go build
	cd utils/simulator && go build -o policies
# cd utils/tfdestroy && go build
	docker run --rm -v $(shell pwd)/utils/simulator/charts:/charts \
	  -v $(shell pwd)/docker:/docs \
	  alpine/helm package /charts/enforcer-sim -d /docs
#	cd utils/collector && go build

# pkg: $(PHONY) tests-pkg simulator-pkg
# 	@echo "Finished packaging."

# docker: build pkg
# docker build -t ${DEV_DOCKER_IMAGE_REPO_BENCHMARK_SUITE}:${DEV_DOCKER_IMAGE_TAG} docker/benchmark-suite
# docker build -t ${DEV_DOCKER_IMAGE_REPO_SIMULATOR}:${DEV_DOCKER_IMAGE_TAG} docker/simulator

# benchmark-suite-docker: build pkg
# docker build -t ${DEV_DOCKER_IMAGE_REPO_BENCHMARK_SUITE}:${DEV_DOCKER_IMAGE_TAG} docker/benchmark-suite

simulator-docker: build simulator-pkg
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

# tests-pkg:
# 	for DIR in $(TESTDIRS); do \
# 		mkdir -p $$DIR ; \
# 		cp tests/control/control $$DIR; \
# 		cp tests/throughput/throughput $$DIR; \
# 		cp tests/latency/latency $$DIR; \
# 		cp tests/windows/windows $$DIR; \
# 		cp tests/enforcer/enforcerbench $$DIR; \
# 		cp utils/collector/collector $$DIR; \
# 		chmod -R +x $$DIR; \
# 	done
# 	for DIR in $(UTILDIRS); do \
# 		mkdir -p $$DIR ; \
# 		cp utils/tfdestroy/tfdestroy $$DIR; \
# 		chmod -R +x $$DIR; \
# 	done

clean:
	rm -rf docker/plan-gen/utils
	rm -rf docker/simulator/utils
# rm -rf docker/benchmark-suite/{tests,utils}
	rm -rf $(ARTIFACTS)
	rm -f docker/enforcer-sim-*.tgz
	rm -rf $(CHARTDIRS)
