FROM golang:1.20-alpine as buildbase

RUN apk add git build-base

WORKDIR /go/src/github.com/Myrtilli/transaction-indexing-svc
COPY vendor .
COPY . .

RUN GOOS=linux go build  -o /usr/local/bin/transaction-indexing-svc /go/src/github.com/Myrtilli/transaction-indexing-svc


FROM alpine:3.9

COPY --from=buildbase /usr/local/bin/transaction-indexing-svc /usr/local/bin/transaction-indexing-svc
RUN apk add --no-cache ca-certificates

ENTRYPOINT ["transaction-indexing-svc"]
