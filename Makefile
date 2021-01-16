GOPATH= $(shell dirname `pwd`)
GIT_VERSION = $(shell git rev-parse --abbrev-ref HEAD) $(shell git rev-parse HEAD)
VERSION=$(shell git rev-parse --short HEAD)
RPM_VERSION=master
PROJECT_NAME  = v2ray-webui
DOCKER        = $(shell which docker)
DOCKER-COMPOSE        = $(shell which docker-compose)
LDFLAGS       = -ldflags "-X 'main.version=\"${RPM_VERSION}-${GIT_VERSION}\"'"
default: build

docker_test:
	echo "not supported for now"
build:
	mkdir -p bin/
	CGO_ENABLED=1 go build ${LDFLAGS} -o bin/$(PROJECT_NAME)
build_arm:
	mkdir -p bin/
	CC=aarch64-linux-gnu-gcc CGO_ENABLED=1 GOARCH=arm64 go build ${LDFLAGS} -o bin/$(PROJECT_NAME)
tar_x86: update_fe_in_repo build
	cp static/geoip.dat bin/
	tar -czvf $(PROJECT_NAME)-$(VERSION).tar.gz bin/ conf/ static/
	rm -rf bin/geoip.dat
tar_arm: update_fe_in_repo build_arm
	cp static/geoip.dat bin/
	tar -czvf $(PROJECT_NAME)-$(VERSION)-arm.tar.gz bin/ conf/ static/
	rm -rf bin/geoip.dat

tar_x86_ci: update_fe_in_repo_ci build
	cp static/geoip.dat bin/
	tar -czvf $(PROJECT_NAME)-master.tar.gz bin/ conf/ static/
	rm -rf bin/geoip.dat

tar_arm_ci: update_fe_in_repo_ci build_arm
	cp static/geoip.dat bin/
	tar -czvf $(PROJECT_NAME)-master-arm.tar.gz bin/ conf/ static/
	rm -rf bin/geoip.dat

tar_all: tar_x86 tar_arm

tar_all_ci: tar_x86_ci tar_arm_ci

build_fe:
	cd front-end
	cd front-end && rm -rf node_modules
	cd front-end && cnpm install
	cd front-end && yarn build

build_fe_ci:
	cd front-end
	cd front-end && rm -rf node_modules
	cd front-end && npm install
	cd front-end && yarn build

update_fe_in_repo: build_fe
	rm -rf static/front-end
	mv front-end/build static/front-end

update_fe_in_repo_ci: build_fe_ci
	rm -rf static/front-end
	mv front-end/build static/front-end
