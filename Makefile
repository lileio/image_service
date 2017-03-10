# vi: ft=make
.PHONY: run proto test benchmark
run:
	go run main.go

proto:
	protoc -I . image_service.proto --go_out=plugins=grpc:$$GOPATH/src

ci:
	go get -v -t ./...
	FILE_LOCATION=../test/out make test
	GOOS=linux GOARCH=amd64 go build -a -o build/image_server ./image_service

test:
	go test -v ./...

benchmark:
	go test -bench=./... -benchmem -benchtime 10s

docker:
	docker build . -t lileio/image_service:latest
