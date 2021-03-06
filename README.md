# aws-vault-proxy

Mimic [aws-vault](https://github.com/99designs/aws-vault) behavior in a Docker container by requesting aws-vault env variables from the host machine and executing a command with the returned aws-vault env variables.

## Build

```
$ ./build.sh
```

## Run

Start proxy server:
```
$ ./bin/aws-vault-proxy-amd64-darwin
```

In separate terminal, start container:
```
$ docker run -it --rm -v `pwd`/bin/aws-vault-proxy-docker-exec-amd64-linux:/usr/local/bin/aws-vault ubuntu
```

Execute command with proxy in container
```
# aws-vault exec PROFILE -- COMMAND
```

## Fake Vault

Build
```
$ GOOS=darwin GOARCH=amd64 go build -o bin/fake-vault ./cmd/fake-vault
```

Start server
```
$ AWS_VAULT_PROXY_COMMAND=./bin/fake-vault go run ./cmd/proxy
```

## Local Development

```
$ AWS_VAULT_PROXY_HOST=127.0.0.1 go run ./cmd/docker-exec exec profile -- ls
```