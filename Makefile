# If the first argument is "run"
ifeq (run,$(firstword $(MAKECMDGOALS)))
  # use the rest as arguments for "run"
  RUN_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # ...and turn them into do-nothing targets
  $(eval $(RUN_ARGS):;@:)
endif

# If the first argument is "dev"
ifeq (dev,$(firstword $(MAKECMDGOALS)))
  # use the rest as arguments for "dev"
  DEV_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # ...and turn them into do-nothing targets
  $(eval $(DEV_ARGS):;@:)
endif

build:
	@go build -o bin/genomdb

run: build
	@./bin/genomdb $(RUN_ARGS)

dev:
	@go run . $(DEV_ARGS)

test:
	@go test -v ./...

# Docker commands
docker-build:
	@docker-compose build

docker-up:
	@docker-compose up -d

docker-down:
	@docker-compose down

docker-logs:
	@docker-compose logs -f

docker-clean:
	@docker-compose down -v
	@docker system prune -f

docker-init:
	@./docker-init.sh

docker-restart:
	@docker-compose restart
