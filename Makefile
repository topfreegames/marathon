PACKAGES = $(shell glide novendor)
GODIRS = $(shell go list ./... | grep -v /vendor/ | sed s@github.com/topfreegames/marathon@.@g | egrep -v "^[.]$$")
PMD = "pmd-bin-5.3.3"

setup:
	@go get -u github.com/Masterminds/glide/...
	@go get -v github.com/spf13/cobra/cobra
	@go get github.com/fzipp/gocyclo
	@go get github.com/topfreegames/goose/cmd/goose
	@go get github.com/fzipp/gocyclo
	@go get github.com/gordonklaus/ineffassign
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

cross:
	@mkdir -p ./bin
	@echo "Building for linux-386..."
	@env GOOS=linux GOARCH=386 go build -o ./bin/marathon-linux-386
	@echo "Building for linux-amd64..."
	@env GOOS=linux GOARCH=amd64 go build -o ./bin/marathon-linux-amd64
	@echo "Building for darwin-386..."
	@env GOOS=darwin GOARCH=386 go build -o ./bin/marathon-darwin-386
	@echo "Building for darwin-amd64..."
	@env GOOS=darwin GOARCH=amd64 go build -o ./bin/marathon-darwin-amd64

install:
	@go install

run:
	@go run main.go start -d -c ./config/local.yaml

build-docker:
	@docker build -t marathon .

run-docker:
	#TODO: REPLACE IP here
	@docker run -i -t --rm -e "MARATHON_REDIS_HOST=10.0.20.81" -p 8080:8080 marathon

kill-redis:
	-redis-cli -p 7575 shutdown

redis: kill-redis
	redis-server ./redis.conf; sleep 1
	redis-cli -p 7575 info > /dev/null

flush-redis:
	redis-cli -p 7575 FLUSHDB

kill-redis-test:
	-redis-cli -p 57575 shutdown

redis-test: kill_redis_test
	redis-server ./redis_test.conf; sleep 1
	redis-cli -p 57575 info > /dev/null

flush-redis-test:
	redis-cli -p 57575 FLUSHDB

test: drop-test db-test
	@go test $(PACKAGES)

coverage: drop-test db-test
	@echo "mode: count" > coverage-all.out
	@$(foreach pkg,$(PACKAGES),\
		go test -coverprofile=coverage.out -covermode=count $(pkg);\
		tail -n +2 coverage.out >> coverage-all.out;)

coverage-html:
	@go tool cover -html=coverage-all.out

static:
	@go vet $(PACKAGES)
	@-gocyclo -over 5 . | egrep -v vendor/
	@#golint
	@for pkg in $$(go list ./... |grep -v /vendor/) ; do \
        golint $$pkg ; \
    done
	@#ineffassign
	@for pkg in $(GODIRS) ; do \
        ineffassign $$pkg ; \
    done
	@${MAKE} pmd

pmd:
	@bash pmd.sh
	@for pkg in $(GODIRS) ; do \
		exclude=$$(find $$pkg -name '*_test.go') && \
		/tmp/pmd-bin-5.4.2/bin/run.sh cpd --minimum-tokens 30 --files $$pkg --exclude $$exclude --language go ; \
    done

pmd-full:
	@bash pmd.sh
	@for pkg in $(GODIRS) ; do \
		/tmp/pmd-bin-5.4.2/bin/run.sh cpd --minimum-tokens 30 --files $$pkg --language go ; \
    done
