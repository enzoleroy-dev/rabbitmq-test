# RabbitMQ Test

## Installation
> Install docker, docker-compose, direnv, go
```sh
$ brew install go@1.24
$ brew install direnv
$ brew install docker docker-compose
```

## For setup package and dependencies
Install go package and setup to `.zshrc` and Execute `$ source .zshrc`
```
# Golang
export PATH=$(go env GOPATH)/bin:$PATH
export GOBIN=$(go env GOPATH)/bin

# Direnv
eval "$(direnv hook zsh)"
```

## Run RabbitMQ
```bash
direnv allow
docker-compose up -d
```

## Terminal 1: Run Consumer
```bash
direnv allow
go run cmd/rabbitmq/main.go
```

## Terminal 2: Run Producer publish message to RabbitMQ
```bash
direnv allow
go run cmd/rabbitmq/example_producer/main.go
```

## Note
- `direnv allow` is export environment variables from `.env` file