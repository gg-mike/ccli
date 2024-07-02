# CI/CD server

CI/CD server is a system agnostic solution which gives users ability to utilise different machines as execution environment, both static (like virtual machines) and dynamic (like containers and pods). It is build using Go language, where goroutines are extensively used to manage two part system (REST server and application engine).

Project was developed in open-source spirit, not only being stored in the public repository, but also using other open-source solutions to build whole environment.

## Local development

### Requirements

- [Docker](https://docs.docker.com/engine/install/)
- [Go](https://go.dev/doc/install)
- [Swag](https://github.com/go-openapi/swag)

### Build

Run `make build` or `go build -o bin/serve main.go` command.

### Start Docker environment

Run `make docker-up` or `docker compose -p ccli -f deployments/docker-compose.yml up -d ` command.

### Migrate

Run `make migrate` or `./bin/ccli migrate` command.

### Start CI/CD server

Run `make run` or `./bin/ccli serve` command.

## Architecture

Detailed architecture of the proposed solution is presented on the diagram below.

![CI/CD architecture](./figures/diagram.svg)
