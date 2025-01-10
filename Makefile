# Based on the Example from Joel Homes, author of "Shipping Go" at
# https://github.com/holmes89/hello-api/blob/main/ch10/Makefile

#  https://stackoverflow.com/a/54776239
SHELL := /bin/bash
GO_VERSION := 1.22  # <1>
COVERAGE_AMT := 70  # should be 80
HEREGOPATH := $(shell go env GOPATH)
CURDIR := $(shell pwd)

# setup: # <2>
# 	install-go
# 	init-go
# 
# install-go: # <3>
# 	wget "https://golang.org/dl/go$(GO_VERSION).linux-amd64.tar.gz"
# 	sudo tar -C /usr/local -xzf go$(GO_VERSION).linux-amd64.tar.gz
# 	rm go$(GO_VERSION).linux-amd64.tar.gz
# 
# init-go: # <4>
#     echo 'export PATH=$$PATH:/usr/local/go/bin' >> $${HOME}/.bashrc
#     echo 'export PATH=$$PATH:$${HOME}/go/bin' >> $${HOME}/.bashrc
# 
# upgrade-go: # <5>
# 	sudo rm -rf /usr/bin/go
# 	wget "https://golang.org/dl/go$(GO_VERSION).linux-amd64.tar.gz"
# 	sudo tar -C /usr/local -xzf go$(GO_VERSION).linux-amd64.tar.gz
# 	rm go$(GO_VERSION).linux-amd64.tar.gz


test:
	go test -v -coverpkg=./... -coverprofile=coverage.out ./... 

coverage-verbose:
	go tool cover -func coverage.out | tee cover.rpt

coverage-ok:
	cat cover.rpt | grep "total:" | awk '{print ((int($$3) > ${COVERAGE_AMT}) != 1) }'

cover-report:
	go tool cover -html=coverage.out -o cover.html

clean:
	rm $$(find . -name "*cover*html" -or -name "*cover.rpt" -or -name "*coverage.out")

check: check-format check-vet test coverage-verbose coverage-ok cover-report lint 

check-format: 
	test -z $$(go fmt ./...)

check-vet: 
	test -z $$(go vet ./...)

testme:
	echo $(HEREGOPATH)

install-lint:
	# https://golangci-lint.run/usage/install/#local-installation to GOPATH
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(HEREGOPATH)/bin v1.57.2
	# report version
	${HEREGOPATH}/bin/golangci-lint --version

lint:
	# golangci-lint run -v ./... 
	${HEREGOPATH}/bin/golangci-lint run ./... 

module-update-tidy:
	go get -u ./...
	go mod tidy

build:
	go build .
