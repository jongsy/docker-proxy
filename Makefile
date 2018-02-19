build:
	docker build -f Dockerfile.build -t dproxy:latest .
	export WORKDIR=/go/src/github.com/jongsy/docker-proxy
	docker create -v /cfg --name go dproxy:latest /bin/true
	docker cp ./ go:/go/src/github.com/jongsy/docker-proxy
	docker run --name builder --volumes-from go -w /go/src/github.com/jongsy/docker-proxy -it dproxy:latest go build -v
	docker cp builder:/go/src/github.com/jongsy/docker-proxy/docker-proxy ./

rebuild:
	docker-compose down
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build docker-proxy.go
	docker-compose build
	./compose-up.sh

proxy:
	docker-compose down
	./compose-up.sh

.PHONY: rebuild proxy
