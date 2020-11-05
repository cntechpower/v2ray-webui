GOPATH= $(shell dirname `pwd`)
GIT_VERSION = $(shell git rev-parse --abbrev-ref HEAD) $(shell git rev-parse HEAD)
VERSION=$(shell git rev-parse --short HEAD)
RPM_VERSION=master
PROJECT_NAME  = api-server
DOCKER        = $(shell which docker)
DOCKER-COMPOSE        = $(shell which docker-compose)
LDFLAGS       = -ldflags "-X 'main.version=\"${RPM_VERSION}-${GIT_VERSION}\"'"
DOCKER_IMAGE  = 10.0.0.2:5000/actiontech/universe-compiler-go1.11-centos6:v2
default: build

docker_test:
	echo "not supported for now"
build:
	mkdir -p bin/
	go build ${LDFLAGS} -o bin/$(PROJECT_NAME)
upload: build
	cp static/geoip.dat bin/
	tar -czvf $(PROJECT_NAME)-$(VERSION).tar.gz bin/
	curl -T  $(PROJECT_NAME)-$(VERSION).tar.gz -u ftp:ftp ftp://10.0.0.2/ci/$(PROJECT_NAME)/
	curl -T  $(PROJECT_NAME)-$(VERSION).tar.gz -u ftp:ftp ftp://10.0.0.2/ci/$(PROJECT_NAME)/$(PROJECT_NAME)-latest.tar.gz
	rm -rf bin/geoip.dat

build_fe:
	cd front-end
	cd front-end && rm -rf node_modules
	cd front-end && cnpm install
	cd front-end && yarn build

update_fe_in_repo: build_fe
	rm -rf static/front-end
	mv front-end/build static/front-end
