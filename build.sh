# macOS
GOOS=darwin GOARCH=amd64 go build -o bin/aws-vault-proxy-amd64-darwin ./cmd/proxy

# Linux
GOOS=linux GOARCH=amd64 go build -o bin/aws-vault-proxy-docker-exec-amd64-linux ./cmd/docker-exec

