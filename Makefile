# vi: ft=make
.PHONY: run proto test benchmark
run:
	go run main.go

proto:
	protoc -I image_service/ image_service/image_service.proto --go_out=plugins=grpc:image_service

test:
	go test -v ./...

benchmark:
	go test -bench=./... -benchmem -benchtime 10s

docker:
	GOOS=linux go build -a -o image_service .
	docker build . -t lileio/image_service:latest
