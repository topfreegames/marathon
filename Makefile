# Copyright (c) 2016 TFG Co <backend@tfgco.com>
# Author: TFG Co <backend@tfgco.com>
#
# Permission is hereby granted, free of charge, to any person obtaining a copy of
# this software and associated documentation files (the "Software"), to deal in
# the Software without restriction, including without limitation the rights to
# use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
# the Software, and to permit persons to whom the Software is furnished to do so,
# subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
# FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
# COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
# IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
# CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

MY_IP=`ifconfig | grep --color=none -Eo 'inet (addr:)?([0-9]*\.){3}[0-9]*' | grep --color=none -Eo '([0-9]*\.){3}[0-9]*' | grep -v '127.0.0.1' | head -n 1`
BIN_PATH = "./bin"
BIN_NAME = "marathon"
GODIRS = $(shell go list . | grep -v /vendor/ | sed s@github.com/topfreegames/marathon@.@g | egrep -v "^[.]$$")

setup-hooks:
	@cd .git/hooks && ln -sf ../../hooks/pre-commit.sh pre-commit

clear-hooks:
	@cd .git/hooks && rm pre-commit

setup: setup-hooks
	@go install github.com/onsi/ginkgo/ginkgo@latest
	@go install github.com/gordonklaus/ineffassign@latest

assets:
	@for pkg in $(GODIRS) ; do \
		go generate -x $$pkg ; \
    done

build:
	@mkdir -p ${BIN_PATH}
	@go build $(GODIRS)
	@go build -o ${BIN_PATH}/${BIN_NAME} main.go

cross: assets
	@mkdir -p ${BIN_PATH}
	@echo "Building for linux-i386..."
	@env GOOS=linux GOARCH=386 go build -o ${BIN_PATH}/${BIN_NAME}-linux-i386
	@echo "Building for linux-x86_64..."
	@env GOOS=linux GOARCH=amd64 go build -o ${BIN_PATH}/${BIN_NAME}-linux-x86_64
	@echo "Building for darwin-i386..."
	@env GOOS=darwin GOARCH=386 go build -o ${BIN_PATH}/${BIN_NAME}-darwin-i386
	@echo "Building for darwin-x86_64..."
	@env GOOS=darwin GOARCH=amd64 go build -o ${BIN_PATH}/${BIN_NAME}-darwin-x86_64
	@chmod +x bin/*

setup-ci:
	@go install github.com/mattn/goveralls@latest
	@go install github.com/onsi/ginkgo/ginkgo@latest

prepare-dev: deps create-db migrate

run:
	@go run main.go start-api -d -c ./config/default.yaml

run-api:
	@go run main.go start-api -d -c ./config/default.yaml

run-workers:
	@go run main.go start-workers -d -c ./config/default.yaml

migrate:
	@go run main.go migrations up

test-migrations:
	@go run main.go migrations up -m test_migrations

drop-db:
	@psql -U postgres -h localhost -p 8585 -f db/drop.sql > /dev/null

create-db:
	@psql -U postgres -h localhost -p 8585 -f db/create.sql > /dev/null

wait-for-pg:
	@until docker exec marathon_postgres_1 pg_isready; do echo 'Waiting for Postgres...' && sleep 1; done
	@sleep 2

deps: start-deps wait-for-pg

start-deps:
	@echo "Starting dependencies using HOST IP of ${MY_IP}..."
	@env MY_IP=${MY_IP} docker-compose --project-name marathon up -d
	@sleep 10
	@echo "Dependencies started successfully."

stop-deps:
	@env MY_IP=${MY_IP} docker-compose --project-name marathon down

test: test-services test-run

test-run:
	@env MY_IP=${MY_IP} ginkgo -r --randomizeAllSpecs --randomizeSuites --cover .
	@$(MAKE) test-coverage-func

test-coverage-func:
	@mkdir -p _build
	@-rm -rf _build/test-coverage-all.out
	@echo "mode: count" > _build/test-coverage-all.out
	@bash -c 'for f in $$(find . -name "*.coverprofile"); do tail -n +2 $$f >> _build/test-coverage-all.out; done'
	@echo
	@echo "=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-"
	@echo "Functions NOT COVERED by Tests"
	@echo "=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-"
	@go tool cover -func=_build/test-coverage-all.out | egrep -v "100.0[%]"

test-coverage: test test-coverage-run

test-coverage-run:
	@mkdir -p _build
	@-rm -rf _build/test-coverage-all.out
	@echo "mode: count" > _build/test-coverage-all.out
	@bash -c 'for f in $$(find . -name "*.coverprofile"); do tail -n +2 $$f >> _build/test-coverage-all.out; done'

test-coverage-html: test-coverage
	@go tool cover -html=_build/test-coverage-all.out

test-coverage-write-html:
	@go tool cover -html=_build/test-coverage-all.out -o _build/test-coverage.html

test-services: stop-deps deps test-db-drop test-db-create test-db-migrate
	@echo "Required test services are up."

test-db-drop:
	@psql -U postgres -h localhost -p 8585 -f db/drop-test.sql > /dev/null

test-db-create:
	@psql -U postgres -h localhost -p 8585 -f db/create-test.sql > /dev/null

test-db-migrate:
	@go run main.go migrations up -c ./config/test.yaml
	@go run main.go migrations up -m test_migrations -c ./config/test.yaml

rtfd:
	@rm -rf docs/_build
	@sphinx-build -b html -d ./docs/_build/doctrees ./docs/ docs/_build/html
	@open docs/_build/html/index.html
