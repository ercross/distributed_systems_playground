# Variables
COMPOSE_FILE := docker-compose.yml
PROJECT_NAME := dsp


.PHONY: build up down restart logs stop clean stats deploy redeploy

build:
	@echo "Building services..."
	docker-compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) build

up:
	@echo "Starting services..."
	docker-compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) up
	@echo "Services started... "

down:
	@echo "Stopping services..."
	docker-compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) down

restart:
	@echo "Restarting services..."
	docker-compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) restart

logs:
	docker-compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) logs -f

stop:
	docker-compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) stop

clean:
	@echo "Cleaning up..."
	docker-compose down -v
	docker system prune -f

stats:
	@echo "Container resource usage:"
	docker -f $(COMPOSE_FILE) -p $(PROJECT_NAME) stats --no-stream

deploy: build up

redeploy: restart logs


# Generate grafana dashboard from spec
generate-dashboard:
	python3 observability/grafana/provisioning/dashboards/builder/compiler.py


# Run k6 test
run-k6-test:
	k6 run test_tools/k6/v0.2.0-saturation-test.js