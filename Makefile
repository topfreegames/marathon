.PHONY: db

MY_IP?=`ifconfig | grep -Eo 'inet (addr:)?([0-9]*\.){3}[0-9]*' | grep -Eo '([0-9]*\.){3}[0-9]*' | grep -v '127.0.0.1' | head -n 1`
CONTAINER_PID:=`docker ps -a | grep marathon | awk ' { print $$1 } '`

setup: setup-global
	@npm install

setup-global:
	@npm install -g nodemon babel-cli webpack mocha bunyan

build:
	@rm -rf lib/
	@webpack

_run-app:
	@nodemon --exec babel-node --presets=es2015 -- src/index.js start | bunyan -o short

run: _run-app

ci-test: ci-test-shutdown docker-build
	@docker-compose -p marathon -f ./docker-compose-ci.yaml run --rm --entrypoint="make _test-unit" app

ci-test-shutdown:
	@-docker-compose -p marathon -f ./docker-compose-ci.yaml stop

# test your application (tests in the test/ directory)
test: _services _test-unit

test-watch: _services
	@env ALLOW_CONFIG_MUTATIONS=true ./node_modules/mocha/bin/mocha --watch --require babel-polyfill --compilers js:babel-core/register 'test/**/*Test.js'

_test-unit: _test-unit-coverage

_test-unit-fast:
	@env ALLOW_CONFIG_MUTATIONS=true ./node_modules/mocha/bin/mocha --require babel-polyfill --compilers js:babel-core/register 'test/unit/**/*Test.js'

_test-unit-watch:
	@env ALLOW_CONFIG_MUTATIONS=true ./node_modules/mocha/bin/mocha --watch --require babel-polyfill --compilers js:babel-core/register 'test/unit/**/*Test.js'

_test-unit-coverage:
	@env ALLOW_CONFIG_MUTATIONS=true ./node_modules/.bin/babel-node node_modules/.bin/babel-istanbul cover node_modules/.bin/_mocha --report text --check-coverage -- -u tdd 'test/unit/**/*Test.js'
	@$(MAKE) _test-unit-coverage-html

_test-unit-coverage-html:
	@env ALLOW_CONFIG_MUTATIONS=true ./node_modules/.bin/babel-node node_modules/.bin/babel-istanbul report --include=./coverage/coverage.json html

static-analysis:
	@./node_modules/.bin/plato -r -e .eslintrc -d report src/
	@open ./report/index.html

_services: _services-shutdown
	@docker-compose -p marathon -f ./docker-compose.yaml up -d

_services-shutdown:
	@docker-compose -p marathon -f ./docker-compose.yaml stop
	@docker-compose -p marathon -f ./docker-compose.yaml rm -f

docker-build:
	@docker build -t marathon .

docker-run: docker-stop
	@docker run -d -t \
		-e "AUTH_USERNAME=auth-username" \
		-e "AUTH_PASSWORD=auth-password" \
		-e "REDIS_URL=//${MY_IP}:6377" \
		-e "PORT=8000" \
		-p 8000:8000 \
		marathon

docker-stop:
	@-echo "Stopping container ${CONTAINER_PID}..." && docker stop ${CONTAINER_PID} && echo "Removing container ${CONTAINER_PID}..." && docker rm ${CONTAINER_PID}

db migrate:
	@psql -h localhost -p 22222 -U postgres -c "SHOW SERVER_VERSION" postgres

drop:
	@psql -h localhost -p 22222 -U postgres -f db/drop.sql > /dev/null
	@echo "Database created successfully!"

db-test migrate-test:
	@psql -h localhost -p 22222 -U postgres -d postgres -c "SHOW SERVER_VERSION"

drop-test:
	@psql -h localhost -p 22222 -U postgres -c "SELECT pg_terminate_backend(pid.pid) FROM pg_stat_activity, (SELECT pid FROM pg_stat_activity where pid <> pg_backend_pid()) pid WHERE datname='khan_test';" postgres
	@psql -h localhost -p 22222 -U postgres -f db/drop-test.sql > /dev/null
	@echo "Test database created successfully!"
