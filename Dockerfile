FROM golang:1.24rc1 AS builder
WORKDIR /tmp/fomo

COPY . /tmp/fomo

RUN go build -o /usr/local/bin/fomo

FROM ubuntu:24.04

RUN apt-get update && apt-get install -y ca-certificates

COPY --from=builder /usr/local/bin/fomo /usr/local/bin/fomo
COPY fomo.json /usr/local/bin/fomo.json

EXPOSE 8080

WORKDIR /usr/local/bin

CMD ["/usr/local/bin/fomo"]