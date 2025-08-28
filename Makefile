.PHONY: integration benchmark run test

run:
	docker-compose up --build --abort-on-container-exit --attach nexus-vector --remove-orphans

bench:
	docker-compose -f docker-compose.benchmark.yml up --build --abort-on-container-exit --attach benchmark-test --remove-orphans

integration:
	docker-compose -f docker-compose.integration.yml up --build --abort-on-container-exit --attach integration-tests --remove-orphans

test:
	go test ./... -v