setup:
	$ make install-air
.PHONY: setup

dev:
	$ air -c .air.toml
.PHONY: dev

start-compose-services:
	$ docker compose up -d --build
PHONY: start-database

stop-compose-services:
	$ docker compose down
PHONY: stop-compose-services

start-dev-compose-services:
	$ docker compose -f docker-compose.dev.yml up -d
PHONY: start-database

stop-dev-compose-services:
	$ docker compose -f docker-compose.dev.yml down
	$ docker volume rm golang-pgx_postgres_data
PHONY: stop-compose-services

install-air:
	go install github.com/cosmtrek/air@latest
	asdf reshim golang
.PHONY: install-air
