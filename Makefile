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

db-test-create:
	@psql -d postgres -f db/drop-test.sql > /dev/null
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

test: run-kafka-zookeeper
	@go test $(PACKAGES)

coverage: run-kafka-zookeeper
	@echo "mode: count" > coverage-all.out
	@$(foreach pkg,$(PACKAGES),\
		go test -coverprofile=coverage.out -covermode=count $(pkg);\
		tail -n +2 coverage.out >> coverage-all.out;)

coverage-html: run-kafka-zookeeper
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

run-kafka-zookeeper: kill-kafka-zookeeper run-zookeeper run-kafka

kill-kafka-zookeeper: kill-kafka kill-zookeeper

run-zookeeper:
	@zookeeper-server-start ./tests/zookeeper.properties 2>&1 > /tmp/marathon-zookeeper.log &

kill-zookeeper:
	@ps aux | egrep "./tests/zookeeper.properties" | egrep -v egrep | awk ' { print $$2 } ' | xargs kill -9
	@rm -rf /tmp/marathon-zookeeper


run-kafka:
	@kafka-server-start ./tests/server.properties 2>&1 > /tmp/marathon-kafka.log &
	@sleep 2
	@kafka-topics --create --partitions 1 --replication-factor 1 --topic test-consumer-1 --zookeeper localhost:3535 2>&1 > /dev/null
	@kafka-topics --create --partitions 1 --replication-factor 1 --topic test-consumer-2 --zookeeper localhost:3535 2>&1 > /dev/null
	@kafka-topics --create --partitions 1 --replication-factor 1 --topic test-consumer-3 --zookeeper localhost:3535 2>&1 > /dev/null
	@kafka-topics --create --partitions 1 --replication-factor 1 --topic test-producer-1 --zookeeper localhost:3535 2>&1 > /dev/null

kill-kafka:
	@ps aux | egrep "./tests/server.properties" | egrep -v egrep | awk ' { print $$2 } ' | xargs kill -9
	@rm -rf /tmp/marathon-kafka-logs
