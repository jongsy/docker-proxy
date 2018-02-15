rebuild:
	docker-compose down
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build docker-proxy.go
	docker-compose build
	./compose-up.sh

proxy:
	docker-compose down
	./compose-up.sh

.PHONY: rebuild proxy
