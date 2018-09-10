FROM alpine:3.7
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
WORKDIR /app
ADD cmd/server/server /app
ADD cmd/goose/goose /app
ADD sqldb/migrations /app/migrations
CMD ["./server"]
