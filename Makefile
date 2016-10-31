PACKAGES = $(shell glide novendor)
DIRS = $(shell find . -type f -not -path '*/\.*' | grep '.go' | grep -v "^[.]\/vendor" | xargs -I {} dirname {} | sort | uniq | grep -v '^.$$')
PMD = "pmd-bin-5.3.3"
MY_IP?=`ifconfig | grep -Eo 'inet (addr:)?([0-9]*\.){3}[0-9]*' | grep -Eo '([0-9]*\.){3}[0-9]*' | grep -v '127.0.0.1' | head -n 1`

setup-hooks:
	@cd .git/hooks && ln -sf ../../hooks/pre-commit.sh pre-commit

setup: setup-hooks
	@go get -u github.com/Masterminds/glide/...
	@go get -u github.com/onsi/ginkgo/ginkgo
	@go get -v github.com/spf13/cobra/cobra
	@go get github.com/fzipp/gocyclo
	@go get github.com/topfreegames/goose/cmd/goose
	@go get github.com/fzipp/gocyclo
	@go get github.com/gordonklaus/ineffassign
	@go get github.com/axw/gocov/gocov
	@go get -u gopkg.in/matm/v1/gocov-html
	@glide install

setup-ci:
	@sudo add-apt-repository -y ppa:masterminds/glide && sudo apt-get update
	@sudo apt-get install -y glide
	@go get github.com/topfreegames/goose/cmd/goose
	@go get github.com/mattn/goveralls
	@glide install

build:
	@go build $(PACKAGES)
	@go build

build-linux-386:
	@mkdir -p ./bin
	@echo "Building for linux-386..."
	@env GOOS=linux GOARCH=386 go build -o ./bin/marathon-linux-386

build-linux-amd64:
	@mkdir -p ./bin
	@echo "Building for linux-amd64..."
	@env GOOS=linux GOARCH=amd64 go build -o ./bin/marathon-linux-amd64

build-darwin-386:
	@mkdir -p ./bin
	@echo "Building for darwin-386..."
	@env GOOS=darwin GOARCH=386 go build -o ./bin/marathon-darwin-386

build-darwin-amd64:
	@mkdir -p ./bin
	@echo "Building for darwin-amd64..."
	@env GOOS=darwin GOARCH=amd64 go build -o ./bin/marathon-darwin-amd64

build-docker:
	@cd ./docker/dev && docker-compose -p marathon_dev build
	@cd ../../

cross: build-linux-386 build-linux-amd64 build-darwin-386 build-darwin-amd64

install:
	@go install

start-dev-dependencies: build-linux-amd64 build-docker
	@MY_IP=${MY_IP} ./docker/dev/start.sh

start-dev: start-dev-dependencies

image: build-linux-amd64
	@docker build -t marathon ./docker

run:
	@go run main.go start-server -d -c ./config/development.yaml

db-test-create:
	@psql -d postgres -f db/drop-test.sql
	@echo "Test database created successfully!"

db-test-migrate:
	@go run main.go migrate -c ./config/test.yaml > /dev/null
	@echo "Test database migrated successfully!"

db-local-create:
	@psql -d postgres -f db/drop.sql > /dev/null
	@echo "Database created successfully!"

db-local-migrate:
	@go run main.go migrate -c ./config/local.yaml
	@echo "Database migrated successfully!"

start-test-dependencies:
	@MY_IP=${MY_IP} ./docker/test/start.sh

test: start-test-dependencies
	@ENV=test ginkgo --cover $(DIRS)

test-verbose: db-test-create db-test-migrate run-kafka-zookeeper
	@ENV=test VERBOSE_TEST=true ginkgo -v --cover $(DIRS)

test-coverage: test
	@rm -rf _build
	@mkdir -p _build
	@echo "mode: count" > _build/test-coverage-all.out
	@bash -c 'for f in $$(find . -name "*.coverprofile"); do tail -n +2 $$f >> _build/test-coverage-all.out; done'

gocov-cover:
	@rm -f _build/test-coverage-all.json _build/http.html
	@gocov convert _build/test-coverage-all.out > _build/test-coverage-all.json
	@gocov-html _build/test-coverage-all.json > _build/http.html
	@open _build/http.html

coverage-html-gocov: test-coverage gocov-cover

coverage-html: test-coverage gocov-cover
	@go tool cover -html=_build/test-coverage-all.out

static:
	@go vet $(PACKAGES)
	@-gocyclo -over 5 . | egrep -v vendor/
	@#golint
	@for pkg in $$(go list ./... |grep -v /vendor/) ; do \
        golint $$pkg ; \
    done
	@#ineffassign
	@for pkg in $(DIRS) ; do \
        ineffassign $$pkg ; \
    done
	@${MAKE} pmd

pmd:
	@bash pmd.sh
	@for pkg in $(DIRS) ; do \
		exclude=$$(find $$pkg -name '*_test.go') && \
		/tmp/pmd-bin-5.4.2/bin/run.sh cpd --minimum-tokens 30 --files $$pkg --exclude $$exclude --language go ; \
    done

pmd-full:
	@bash pmd.sh
	@for pkg in $(DIRS) ; do \
		/tmp/pmd-bin-5.4.2/bin/run.sh cpd --minimum-tokens 30 --files $$pkg --language go ; \
    done
