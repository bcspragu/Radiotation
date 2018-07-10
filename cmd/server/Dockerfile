FROM alpine:3.7
RUN apk update && apk add ca-certificates
WORKDIR /server
ADD server server
CMD ["./server"]

