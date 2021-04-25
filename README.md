# aws-vault-proxy

Mimic [aws-vault](https://github.com/99designs/aws-vault) behavior in a Docker container by requesting aws-vault env variables from the host machine and executing a command with the returned aws-vault env variables.

## Build

```
$ ./build.sh
```

## Run

Start proxy server
```
$ ./bin/aws-vault-proxy-amd64-darwin server
```

Start container (in separate terminal)
```
$ docker run -it --rm -v `pwd`/bin/aws-vault-proxy-amd64-linux:/usr/local/bin/aws-vault ubuntu
```

Execute command with proxy in container
```
# aws-vault exec PROFILE -- COMMAND
```
