wait-for-pg:
	@until docker exec marathon_postgres_1 pg_isready; do echo 'Waiting for Postgres...' && sleep 1; done
	@sleep 2

deps: start-deps wait-for-pg

start-deps:
	@docker-compose --project-name marathon up -d

stop-deps:
	@docker-compose --project-name marathon down
