FROM alpine:latest

RUN apk add --no-cache ca-certificates
ADD image_service /
CMD ["/image_service"]
