FROM golang:1.24rc1 AS builder
WORKDIR /tmp/fomo

COPY . /tmp/fomo

RUN go build -o /app/fomo/fomo

FROM ubuntu:24.04

RUN apt-get update && apt-get install -y ca-certificates

RUN mkdir -p /app/fomo/data
COPY --from=builder /app/fomo/fomo /app/fomo/fomo
COPY config.json /app/fomo/config.json

EXPOSE 8080

ENV FOMO_CONFIG=/app/fomo/config.json
ENV FOMO_DATA_DIR=/app/fomo/data

ENTRYPOINT ["/app/fomo/fomo"]