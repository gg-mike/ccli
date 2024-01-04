APP=ccli
CMD=serve
# CMD=migrate

.PHONY: build
build: fmt
	go build -o bin/${APP} main.go

.PHONY: run
run: build
	./bin/${APP} ${CMD}

.PHONY: migrate
migrate: build
	./bin/${APP} migrate

.PHONY: docs
docs:
	swag init -o ./docs -pd

.PHONY: clean
clean:
	go clean
	rm -rf bin && mkdir bin

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: test
test:
	go test -v ./pkg/...

.PHONY: docker-up
docker-up:
	docker compose -p ccli -f deployments/docker-compose.yml up -d 

.PHONY: docker-down
docker-down:
	docker compose -p ccli -f deployments/docker-compose.yml down