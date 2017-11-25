#
# Makefile for 'playground'
#
# generates the html describing the API's
#

NAME	:= playground
DESC	:= "various golang thingee's"
PREFIX	?= usr/local
VERSION := $(shell git describe --tags --always --dirty)
GOVERSION := $(shell go version)
BUILDTIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILDDATE := $(shell date -u +"%B %d, %Y")
BUILDER	:= $(shell echo "`git config user.name` <`git config user.email`>")
BUILD_NUMBER_FILE=.buildnum
BUILD_NUM := $(shell if [ -f ${BUILD_NUMBER_FILE} ]; then cat ${BUILD_NUMBER_FILE}; else echo 0; fi)
PKG_RELEASE ?= 1
PROJECT_URL := "git@github.com:mchudgins/playground.git"
#HYGIENEPKG := "github.com/mchudgins/certMgr/pkg/utils"
#LDFLAGS	:= -X '$(HYGIENEPKG).version=$(VERSION)' \
#	-X '$(HYGIENEPKG).buildTime=$(BUILDTIME)' \
#	-X '$(HYGIENEPKG).builder=$(BUILDER)' \
#	-X '$(HYGIENEPKG).goversion=$(GOVERSION)' \
#	-X '$(HYGIENEPKG).buildNum=$(BUILD_NUM)'

DEPS := $(shell ls *.go | sed 's/.*_test.go//g')

container: $(DEPS) docker/Dockerfile $(GENERATED_FILES)
	CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags "-s $(LDFLAGS)" -o bin/$(NAME)
	@-rm docker/app
#	upx -9 -q bin/$(NAME) -o docker/app
	cp bin/$(NAME) docker/playground
	sudo docker build -t mchudgins/$(NAME):$(BUILD_NUM) docker
	sudo docker tag mchudgins/$(NAME):$(BUILD_NUM) mchudgins/$(NAME):latest

pkg/cmd/backend/htmlGen/assets.go: pkg/cmd/backend/htmlGen/test.yaml pkg/cmd/backend/htmlGen/defaultHTML.go
	go run main.go htmlGen pkg/cmd/backend/htmlGen/test.yaml >pkg/cmd/backend/htmlGen/apiList.html
	staticfiles -o pkg/cmd/backend/htmlGen/assets.go -exclude '*.yaml,*.go' pkg/cmd/backend/htmlGen

pkg/cmd/backend/assets.go: pkg/cmd/backend/assets/service.swagger.json
	staticfiles -o pkg/cmd/backend/assets.go pkg/cmd/backend/assets

fmt:
	-gometalinter .
	-goimports -w .

run: pkg/cmd/backend/htmlGen/assets.go pkg/cmd/backend/assets.go
	go run main.go

clean: fmt
	@-rm pkg/cmd/backend/htmlGen/assets.go

coverage:
	go test -coverprofile=/tmp/c.out ./echo
	go tool cover -html /tmp/c.out
