FROM golang:1.22.1-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN make proxyserver

FROM alpine:latest

RUN adduser -D kengrokuser

WORKDIR /app

COPY --from=builder /app/bin/* /app/

USER kengrokuser

ENTRYPOINT ["/app/proxyserver", "3000"]
